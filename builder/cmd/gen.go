package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	gencmd "github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
	"github.com/jessevdk/go-flags"
)

//go:embed templ/templates/*
var et embed.FS

func generate() error {

	err := writeTemplates()
	if err != nil {
		return err
	}

	// templateDir := filepath.Join(fnDir, "build/templates")
	swaggerFile := filepath.Join(fnDir, "swagger.yaml")
	targetDir := filepath.Join(fnDir, "build/app")

	err = os.MkdirAll(targetDir, 0777)
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
	// m.SkipOperations = false

	m.ServerPackage = "restapi"

	m.Models.ModelPackage = "models"

	err = m.Execute([]string{})
	if err != nil {
		return err
	}

	return nil
}

func writeTemplates() error {

	templateDir := filepath.Join(fnDir, "build/templates")

	log.Printf("deleting templates in %s\n", templateDir)
	err := os.RemoveAll(templateDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(templateDir, 0777)
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

	// fullDir := dir

	fmt.Printf("DIRECTORY1 %v\n", dir)

	// create target directory
	targetDir := strings.Replace(dir, "templ", fmt.Sprintf("%s/build", fnDir), 1)
	err := os.MkdirAll(targetDir, 0777)
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

		fmt.Printf("NAME %v\n", file.Name())

		if file.IsDir() {

			err = writeDirFiles(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}

		} else {

			fmt.Printf("FROM %v TO %v\n", filepath.Join(dir, file.Name()), filepath.Join(targetDir, file.Name()))

			err := writeFile(filepath.Join(dir, file.Name()),
				filepath.Join(targetDir, file.Name()))
			if err != nil {
				fmt.Printf("!!!!!!!!!!!!!!!!1 ERR %v\n", err)
			}

		}

	}

	return nil

}
