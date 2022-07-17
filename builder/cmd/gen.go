package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

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

	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}

	m := &gencmd.Server{}
	m.Shared.Spec = flags.Filename(swaggerFile)
	// m.Shared.Target = flags.Filename(targetDir)
	m.Shared.ConfigFile = flags.Filename(filepath.Join(fnDir, "build/templates/server.yaml"))

	// m.SkipModels = false
	// m.SkipOperations = false

	m.ServerPackage = "restapi"

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

		fin, err := et.ReadFile(filepath.Join("templ/templates", file.Name()))
		if err != nil {
			return err
		}

		log.Printf("writing %s\n", file.Name())
		err = os.WriteFile(filepath.Join(fnDir, "build/templates", file.Name()), fin, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
