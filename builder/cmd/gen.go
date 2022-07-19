package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	direktivmodel "github.com/direktiv/direktiv/pkg/model"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	gencmd "github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

//go:embed templ/templates/*
var et embed.FS

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

func generate() error {

	var err error

	fnDir, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}

	log.Printf("using directory '%s'\n", fnDir)

	err = writeTemplates()
	if err != nil {
		return err
	}

	swaggerFile := filepath.Join(fnDir, "swagger.yaml")
	targetDir := filepath.Join(fnDir, "build/app")

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}

	gomod := []byte(`module app

go 1.18`)

	os.WriteFile(filepath.Join(targetDir, "go.mod"), gomod, 0644)

	m := &gencmd.Server{}
	m.Shared.Spec = flags.Filename(swaggerFile)
	m.Shared.Target = flags.Filename(targetDir)
	m.Shared.ConfigFile = flags.Filename(filepath.Join(fnDir, "build/templates/server.yaml"))

	m.Shared.TemplateDir = flags.Filename(filepath.Join(fnDir, "build/templates"))
	m.Shared.AllowTemplateOverride = true

	m.SkipModels = false
	m.ServerPackage = "restapi"
	m.Models.ModelPackage = "models"

	err = m.Execute([]string{})
	if err != nil {
		return err
	}

	err = writeTests()
	if err != nil {
		return err
	}

	return nil
}

func writeTests() error {

	swaggerFile := filepath.Join(fnDir, "swagger.yaml")

	specDoc, err := loads.Spec(swaggerFile)
	if err != nil {
		return err
	}

	// get version
	version := specDoc.Spec().Info.Version
	title := specDoc.Spec().Info.Title

	// create test dir for this version
	testPath := filepath.Join(fnDir, "tests",
		fmt.Sprintf("version-%s", version))
	err = os.MkdirAll(testPath, 0755)
	if err != nil {
		return err
	}

	log.Printf("generating tests for %s version %s", title, version)

	paths := specDoc.Spec().Paths
	post := paths.Paths["/"].Post

	fn := post.Extensions["x-direktiv-function"]

	// create new workflow
	var workflow direktivmodel.Workflow

	// add function
	var fd direktivmodel.ReusableFunctionDefinition

	// good enough to remove the stuf we don't need
	fnString := strings.Replace(fn.(string), "- ", "  ", 1)
	fnString = strings.Replace(fnString, "functions:", "", 1)
	err = yaml.Unmarshal([]byte(fnString), &fd)
	if err != nil {
		return err
	}

	workflow.Functions = []direktivmodel.FunctionDefinition{
		&fd,
	}

	// add states for each example
	examples := post.Extensions["x-direktiv-examples"].([]interface{})

	// init state
	workflow.States = make([]direktivmodel.State, 0)

	for a := range examples {
		ex := examples[a].(map[string]interface{})

		// need to remove the list '-'
		state := strings.Replace(ex["content"].(string), "- ", "  ", 1)

		var action direktivmodel.ActionState
		yaml.Unmarshal([]byte(state), &action)

		// assign new ids and link them up
		action.ID = fmt.Sprintf("state%d", a)
		if a+1 != len(examples) {
			action.Transition = fmt.Sprintf("state%d", a+1)
		}

		workflow.States = append(workflow.States, &action)
	}

	err = writeEventTest(workflow, testPath, version, title)
	if err != nil {
		return err
	}

	return writeTestFile(filepath.Join(testPath, "tests.yaml"), workflow)

}

func writeEventTest(workflow direktivmodel.Workflow, testPath, version,
	title string) error {

	start := &direktivmodel.EventStart{
		Event: &direktivmodel.StartEventDefinition{
			// Type: "io.direktiv.function.test",
			Type: "io.direktiv.function.test",
			Context: map[string]interface{}{
				"function": title,
				"version":  version,
			},
		},
	}
	start.Type = direktivmodel.StartTypeEvent

	workflow.Start = start

	fmt.Println(start)

	return writeTestFile(filepath.Join(testPath, "tests-event.yaml"), workflow)

	// Type    string                 `yaml:"type"`
	// Context map[string]interface{} `yaml:"context,omitempty"`

	// return nil

}

func writeTestFile(file string, data interface{}) error {

	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	wf, err := os.Create(file)
	if err != nil {
		return err
	}
	defer wf.Close()

	_, err = wf.Write(out)
	return err

}

func writeTemplates() error {

	templateDir := filepath.Join(fnDir, "build/templates")

	log.Printf("deleting templates in %s\n", templateDir)
	err := os.RemoveAll(templateDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(templateDir, 0755)
	if err != nil {
		return err
	}

	// write templates
	e, err := et.ReadDir("templ/templates")
	if err != nil {
		return err
	}

	for i := range e {

		file := e[i]

		if file.IsDir() {

			err = writeDirFiles(filepath.Join("templ/templates", file.Name()))
			if err != nil {
				return err
			}

		} else {

			err := writeFile(filepath.Join("templ/templates", file.Name()),
				filepath.Join(fnDir, "build/templates", file.Name()))
			if err != nil {
				return err
			}

		}

	}

	return nil
}

func writeFile(src, target string) error {

	fin, err := et.ReadFile(src)
	if err != nil {
		return err
	}

	log.Printf("writing %s\n", target)
	err = os.WriteFile(target, fin, 0644)
	if err != nil {
		return err
	}

	return nil
}

func writeDirFiles(dir string) error {

	// create target directory
	targetDir := strings.Replace(dir, "templ", fmt.Sprintf("%s/build", fnDir), 1)
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
		return err
	}

	e, err := et.ReadDir(dir)
	if err != nil {
		return err
	}

	for i := range e {

		file := e[i]
		if file.IsDir() {

			err = writeDirFiles(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}

		} else {

			err := writeFile(filepath.Join(dir, file.Name()),
				filepath.Join(targetDir, file.Name()))
			if err != nil {
				return err
			}

		}

	}

	return nil

}
