package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func generateVerifications(verifications []Verification) error {
	tmpl, err := template.ParseFiles("internal/codegen/templates/verification.go.tmpl")
	if err != nil {
		return err
	}

	for _, verification := range verifications {
		fmt.Printf("generating source verification \"%s\"\n", verification.NameSnake)

		var buf bytes.Buffer
		// Execute the template and write the output to the buffer
		err = tmpl.Execute(&buf, verification)
		if err != nil {
			fmt.Printf("Failed to execute template: %v\n", err)
			return err
		}
		writeFile("verification_"+strings.ReplaceAll(verification.NameSnake, "_", "")+".go", buf.String())
	}

	return nil
}

func generateModel(verifications []Verification) error {
	fmt.Println("generating \"model.go\" file")

	tmpl, err := template.ParseFiles("internal/codegen/templates/model.go.tmpl")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	// Execute the template and write the output to the buffer
	err = tmpl.Execute(&buf, verifications)
	if err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		return err
	}
	writeFile("model.go", buf.String())

	return nil
}

func writeFile(fileName string, content string) {
	// Define the directory and file name
	dir := "internal/provider/sourceverification/generated"

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
