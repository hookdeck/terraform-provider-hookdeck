package main

import (
	"log"
	codegen "terraform-provider-hookdeck/internal/codegenv2"
)

func main() {
	log.Println("Generating source verification codes... should NOT")
	if err := codegen.RunCodeGen(); err != nil {
		log.Panicln(err)
	}
}

// Run "go generate" to generate Hookdeck source verification codes

////go:generate go run .
////go:generate gofmt -w ../../internal/provider/sourceverification/generated
