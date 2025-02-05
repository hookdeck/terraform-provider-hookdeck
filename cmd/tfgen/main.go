package main

import (
	"fmt"
	"os"
	"terraform-provider-hookdeck/internal/tfgen"
)

func main() {
	fmt.Println("generate TF Provider Code Specification")
	if err := tfgen.Generate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

//go:generate go run .

//go:generate tfplugingen-framework generate resources --input ../../assets/provider-code-spec.json --output ../../internal/generated/tfplugingen
