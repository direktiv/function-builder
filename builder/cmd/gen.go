package main

import (
	"embed"
	"errors"
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

//go:embed templ/workflows/*
var wfTempl embed.FS

//go:embed mod/gomod
var gomod string

//go:embed mod/gosum
var gosum string

var karateImage = "gcr.io/direktiv/functions/karate:1.0"

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

	if _, err := os.Stat(filepath.Join(targetDir, "go.mod")); errors.Is(err, os.ErrNotExist) {
		os.WriteFile(filepath.Join(targetDir, "go.mod"), []byte(gomod), 0644)
		os.WriteFile(filepath.Join(targetDir, "go.sum"), []byte(gosum), 0644)
	}

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

	fnName = title

	// create test dir for this version
	testPath := filepath.Join(fnDir, "tests",
		fmt.Sprintf("v%s", version))
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

	// replace version with 'test'
	fnString = strings.Replace(fnString,
		fmt.Sprintf(":%s", version), ":test", 1)

	err = yaml.Unmarshal([]byte(fnString), &fd)
	if err != nil {
		return err
	}

	workflow.Functions = []direktivmodel.FunctionDefinition{
		&fd,
	}

	// add states for each example

	var examples []interface{}
	if post.Extensions["x-direktiv-examples"] != nil {
		examples = post.Extensions["x-direktiv-examples"].([]interface{})
	}

	// init state
	workflow.States = make([]direktivmodel.State, 0)

	for a := range examples {
		ex := examples[a].(map[string]interface{})

		// split at id, we have to use strings because there is no toYAML in markdown conversion
		states := strings.Split(ex["content"].(string), "- id:")

		for i := range states {

			if states[i] != "" {
				s := fmt.Sprintf("  id: %v", states[i])

				var action direktivmodel.ActionState
				err := yaml.Unmarshal([]byte(s), &action)
				if err != nil {
					return err
				}

				// 0th item is split before id:
				if i == 1 {
					action.ID = fmt.Sprintf("state%d", a)
				}

				// last item transitions to next example
				if a+1 != len(examples) && i+1 == len(states) {
					action.Transition = fmt.Sprintf("state%d", a+1)
				}

				workflow.States = append(workflow.States, &action)
			}

		}

	}

	err = writeEventTest(testPath, version, title)
	if err != nil {
		return err
	}

	secrets, ok := post.Extensions["x-direktiv-secrets"].([]interface{})
	if !ok {
		secrets = []interface{}{}
	}

	err = writeKarateTest(testPath, secrets, version)
	if err != nil {
		return err
	}

	return writeTestFile(filepath.Join(testPath, "tests.yaml"), workflow)

}

func writeTestFeatureFile(testPath string, secrets []string) error {

	var secretsString strings.Builder

	for a := range secrets {

		secretsString.WriteString(
			fmt.Sprintf("* def %s = karate.properties['%s']\n",
				secrets[a], secrets[a]))

	}

	testFeature := `
Feature: Basic

# The secrects can be used in the payload with the following syntax #(mysecretname)
Background:
SECRETS

Scenario: get request

	Given url karate.properties['testURL']

	And path '/'
	And header Direktiv-ActionID = 'development'
	And header Direktiv-TempDir = '/tmp'
	And request
	"""
	{
		"commands": [
		{
			"command": "ls -la",
			"silent": true,
			"print": false,
		}
		]
	}
	"""
	When method POST
	Then status 200
	And match $ ==
	"""
	{
	"APPNAME": [
	{
		"result": "#notnull",
		"success": true
	}
	]
	}
	"""
	`

	testFeature = strings.Replace(testFeature, "SECRETS",
		secretsString.String(), 1)
	testFeature = strings.Replace(testFeature, "APPNAME",
		fnName, 1)

	testFeatureFile := filepath.Join(testPath, "karate.yaml.test.feature")
	_, err := os.Stat(testFeatureFile)
	if err == nil {
		return nil
	}

	os.WriteFile(testFeatureFile, []byte(testFeature), 0644)

	return nil

}

func writeKarateTest(testPath string, secrets []interface{}, version string) error {

	// get secret names list
	secretStrings := []string{}
	for a := range secrets {
		s := secrets[a].(map[string]interface{})
		secretName := s["name"]
		secretStrings = append(secretStrings, secretName.(string))
	}

	// write test.feature file
	err := writeTestFeatureFile(testPath, secretStrings)
	if err != nil {
		return err
	}

	// write workflow
	var workflow direktivmodel.Workflow
	workflow.States = make([]direktivmodel.State, 0)

	var fd direktivmodel.ReusableFunctionDefinition
	fd.Image = karateImage
	fd.Type = direktivmodel.ReusableContainerFunctionType
	fd.ID = "karate"

	// add karate tester to the workflow
	workflow.Functions = []direktivmodel.FunctionDefinition{
		&fd,
	}

	input := make(map[string]interface{})

	var command1 strings.Builder
	command1.WriteString("java -DtestURL=jq(.host)")
	for a := range secretStrings {
		command1.WriteString(fmt.Sprintf(" -D%s=jq(.secrets.%s)",
			secretStrings[a], secretStrings[a]))
	}
	command1.WriteString(" -jar /karate.jar test.feature")

	command2 := "cat target/karate-reports/karate-summary-json.txt"

	input["commands"] = []map[string]interface{}{
		{
			"command": command1.String(),
			"print":   false,
		},
		{
			"command": command2,
		},
	}

	input["logging"] = "info"

	// add actual test
	var action direktivmodel.ActionState
	action.ID = "run-test"
	action.Type = direktivmodel.StateTypeAction

	action.Action = &direktivmodel.ActionDefinition{
		Function: "karate",
		Input:    input,
		Secrets:  secretStrings,
		Files: []direktivmodel.FunctionFileDefinition{
			{
				Key:   "test.feature",
				Scope: "workflow",
			},
		},
	}

	workflow.States = append(workflow.States, &action)

	// only write test script if it does not exist
	if _, err := os.Stat(filepath.Join(fnDir, "run-tests.sh")); errors.Is(err, os.ErrNotExist) {
		err = writeKarateTestScript(secretStrings, version)
		if err != nil {
			return err
		}
	}

	return writeTestFile(filepath.Join(testPath, "karate.yaml"), workflow)
}

func writeKarateTestScript(secrets []string, version string) error {

	var sb strings.Builder
	var karateArgs strings.Builder

	sb.WriteString("#!/bin/bash\n\n")

	sb.WriteString("if [[ -z \"${DIREKTIV_TEST_URL}\" ]]; then\n")
	sb.WriteString("	echo \"Test URL is not set, setting it to http://localhost:9191\"\n")
	sb.WriteString("	DIREKTIV_TEST_URL=\"http://localhost:9191\"\n")
	sb.WriteString("fi\n\n")

	for a := range secrets {
		secret := secrets[a]
		sb.WriteString(fmt.Sprintf("if [[ -z \"${DIREKTIV_SECRET_%s}\" ]]; then\n", secret))
		sb.WriteString(fmt.Sprintf("	echo \"Secret %s is required, set it with DIREKTIV_SECRET_%s\"\n",
			secret, secret))
		sb.WriteString("	exit 1\n")
		sb.WriteString("fi\n\n")

		karateArgs.WriteString(fmt.Sprintf("-D%s=\"${DIREKTIV_SECRET_%s}\" ", secret, secret))
	}

	cmd := fmt.Sprintf("docker run --network=host -v `pwd`/tests/:/tests direktiv/karate "+
		"java -DtestURL=${DIREKTIV_TEST_URL} -Dlogback.configurationFile=/logging.xml %s "+
		"-jar /karate.jar /tests/v%s/karate.yaml.test.feature ${*:1}", karateArgs.String(), version)

	sb.WriteString(cmd)

	return os.WriteFile(filepath.Join(fnDir, "run-tests.sh"), []byte(sb.String()), 0755)

}

func writeEventTest(testPath, version, title string) error {

	repl := func(path, target string) error {
		fin, err := wfTempl.ReadFile(path)
		if err != nil {
			return err
		}

		wf := strings.ReplaceAll(string(fin), "VERSION", version)
		wf = strings.ReplaceAll(wf, "APPNAME", title)

		return os.WriteFile(target, []byte(wf), 0644)

	}

	err := repl("templ/workflows/test-event.yaml",
		filepath.Join(testPath, "tests-event.yaml"))
	if err != nil {
		return err
	}

	return repl("templ/workflows/karate-event.yaml",
		filepath.Join(testPath, "karate-event.yaml"))

}

func writeTestFile(file string, data interface{}) error {

	_, err := os.Stat(file)
	if err == nil {
		return nil
	}

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
