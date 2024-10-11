package main

import (
	"log"
	"terraform-provider-hookdeck/internal/codegen"
)

func main() {
	if err := codegen.RunCodeGen(); err != nil {
		log.Panicln(err)
	}
}

// Run "go generate" to generate Hookdeck source verification codes

//go:generate go run .
//go:generate gofmt -w ../../internal/provider/sourceverification/generated
