package codegenv2

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

func generateSources(sourceTypes []SourceType) error {
	// Load all templates
	tmpl, err := template.ParseFiles(
		getRelativePath("templates/source.go.tmpl"),
		getRelativePath("templates/source_config_resource.go.tmpl"),
		getRelativePath("templates/resource.go.tmpl"),
	)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	// Generate resource.go
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "resource.go.tmpl", nil)
	if err != nil {
		return fmt.Errorf("failed to execute resource template: %w", err)
	}
	writeFile("resource.go", buf.String())

	// Generate source files
	for _, sourceType := range sourceTypes {
		fmt.Printf("generating source \"%s\"\n", sourceType.NameSnake)

		buf.Reset()
		err = tmpl.ExecuteTemplate(&buf, "source.go.tmpl", sourceType)
		if err != nil {
			fmt.Printf("Failed to execute template: %v\n", err)
			return err
		}
		writeFile("source_"+strings.ReplaceAll(sourceType.NameSnake, "_", "")+".go", buf.String())
	}

	return nil
}

func writeFile(fileName string, content string) {
	// Define the directory and file name
	dir := getRelativePath("../provider/sourcev2/generated")

	// Create the "generated" directory if it doesn't exist
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return
	}

	// Define the full path to the file
	filePath := filepath.Join(dir, fileName)

	// Open the file for writing, create it if it doesn't exist
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("Failed to write to file: %v\n", err)
		return
	}
}

// getRelativePath returns the absolute path relative to the current file.
func getRelativePath(relativePath string) string {
	_, filename, _, _ := runtime.Caller(1)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, relativePath)
}
