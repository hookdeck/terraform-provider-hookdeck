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
