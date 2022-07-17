package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var fnName, fnDir string

type genFile struct {
	name    string
	mode    fs.FileMode
	replace bool
	rename  string
}

//go:embed templ/project/*
var ef embed.FS

func prepDir() error {

	if fnDir == "" {
		log.Printf("directory note set\n")
		fnDir = fnName
	}

	log.Printf("using directory '%s'\n", fnDir)

	// create dir
	err := os.MkdirAll(fnDir, 0755)
	if err != nil {
		return err
	}

	// check if it is empty
	e, err := os.ReadDir(fnDir)
	if err != nil {
		return err
	}

	if len(e) > 0 {
		errMsg := fmt.Sprintf("target directory %s not empty", fnDir)
		log.Println(errMsg)
		return fmt.Errorf(errMsg)
	}

	return nil

}

func copyFiles() error {

	// writes the file as a base name to the directory
	var writeFile = func(file genFile) error {
		fin, err := ef.ReadFile(file.name)
		if err != nil {
			return err
		}

		// replace APPNAME with the name of the function
		if file.replace {
			r := strings.ReplaceAll(string(fin), "APPNAME", fnName)
			fin = []byte(r)
		}

		// if rename, change name
		if file.rename != "" {
			file.name = file.rename
		}

		return os.WriteFile(filepath.Join(fnDir, filepath.Base(file.name)),
			fin, file.mode)

	}

	files := []genFile{
		{"templ/project/LICENSE", 0644, false, ""},
		{"templ/project/gitignore", 0644, false, ".gitignore"},
		{"templ/project/run.sh", 0755, true, ""},
		{"templ/project/swagger.yaml", 0644, true, ""},
		{"templ/project/Dockerfile", 0644, true, ""},
	}

	for f := range files {
		file := files[f]
		err := writeFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepare() error {

	err := prepDir()
	if err != nil {
		return err
	}

	err = copyFiles()
	if err != nil {
		return err
	}

	log.Printf("function %s prepared in %s\n", fnName, fnDir)

	return nil
}
