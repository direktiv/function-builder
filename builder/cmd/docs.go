package main

import (
	"log"
	"os"
	"path/filepath"

	genmark "github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
	"github.com/jessevdk/go-flags"
)

func docs() error {

	var err error

	fnDir, err = os.Getwd()
	if err != nil {
		log.Println(err)
	}

	swaggerFile := filepath.Join(fnDir, "swagger.yaml")
	readmeFile := filepath.Join(fnDir, "readme.md")

	m := &genmark.Markdown{}
	m.Shared.Spec = flags.Filename(swaggerFile)

	m.Shared.WithFlatten = []string{"full"}
	m.Output = flags.Filename(readmeFile)

	m.Shared.Target = flags.Filename(readmeFile)
	m.Shared.TemplateDir = flags.Filename(filepath.Join(fnDir, "build/templates"))
	m.Shared.AllowTemplateOverride = true

	return m.Execute([]string{})
}
