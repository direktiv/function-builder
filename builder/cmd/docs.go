package main

import (
	"fmt"
	"path/filepath"

	genmark "github.com/go-swagger/go-swagger/cmd/swagger/commands/generate"
	"github.com/jessevdk/go-flags"
)

func docs() error {

	swaggerFile := filepath.Join(fnDir, "swagger.yaml")
	readmeFile := filepath.Join(fnDir, "readme.md")

	m := &genmark.Markdown{}

	fmt.Printf("M %v", m)

	m.Shared.Spec = flags.Filename(swaggerFile)
	// m.Output = flags.Filename(readmeFile)

	m.Shared.WithFlatten = []string{"full"}
	m.Output = flags.Filename(readmeFile)

	m.Shared.Target = flags.Filename(readmeFile)
	m.Shared.TemplateDir = flags.Filename(filepath.Join(fnDir, "build/templates"))
	m.Shared.AllowTemplateOverride = true
	// swagger generate markdown -f /tmp/app/swagger.yaml --output=/tmp/app/readme.md -t /tmp/app/ --template-dir=templates/ --with-flatten=full

	return m.Execute([]string{})
}
