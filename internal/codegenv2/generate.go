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
	// Load templates
	tmpl := template.New("").Funcs(template.FuncMap{
		"eq": func(a, b interface{}) bool {
			fmt.Printf("eq: comparing %v with %v\n", a, b)
			return a == b
		},
		"getFieldType": func(f FieldType) string {
			t := f.getFieldType()
			fmt.Printf("getFieldType: got %s\n", t)
			return t
		},
		"asStringField": func(f FieldType) StringField {
			fmt.Printf("asStringField: type=%T\n", f)
			if sf, ok := f.(StringField); ok {
				return sf
			}
			return StringField{}
		},
		"asArrayField": func(f FieldType) ArrayField {
			fmt.Printf("asArrayField: type=%T\n", f)
			if af, ok := f.(ArrayField); ok {
				return af
			}
			return ArrayField{}
		},
		"asObjectField": func(f FieldType) ObjectField {
			fmt.Printf("asObjectField: type=%T\n", f)
			if of, ok := f.(ObjectField); ok {
				return of
			}
			return ObjectField{}
		},
	})

	// Debug: Print source types
	for _, sourceType := range sourceTypes {
		fmt.Printf("Source type %s fields:\n", sourceType.NamePascal)
		for _, field := range sourceType.Fields {
			fmt.Printf("  - %s: Type=%T\n", field.NamePascal, field.Type)
		}
	}

	// Parse helpers first
	helpersContent, err := os.ReadFile(getRelativePath("templates/helpers.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read helpers template: %w", err)
	}
	_, err = tmpl.Parse(string(helpersContent))
	if err != nil {
		return fmt.Errorf("failed to parse helpers template: %w", err)
	}

	// Parse main template
	content, err := os.ReadFile(getRelativePath("templates/source_config_resource.go.tmpl"))
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	_, err = tmpl.New("source_config_resource.go.tmpl").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Generate source files
	var buf bytes.Buffer
	for _, sourceType := range sourceTypes {
		fmt.Printf("generating source \"%s\"\n", sourceType.NameSnake)

		buf.Reset()
		err = tmpl.ExecuteTemplate(&buf, "source_config_resource.go.tmpl", sourceType)
		if err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", sourceType.NameSnake, err)
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
