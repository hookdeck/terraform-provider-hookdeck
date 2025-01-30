package codegen

import (
	"errors"
	"log"
	"sort"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

type VerificationProperty struct {
	NameSnake      string
	NameCamel      string
	NamePascal     string
	Optional       bool
	Required       bool
	Sensitive      bool
	PointerString  string // either empty string or "Pointer" if property is optional
	TypeString     string // "String" | "Float" ...
	IsEnum         bool
	EnumNameString string
}

type Verification struct {
	Ref        string
	ConfigRef  string
	NameSnake  string
	NameCamel  string
	NamePascal string
	NameConfig string // similar to NamePascal but for ThreeDEye it's 3DEye and other casing diff
	Properties []VerificationProperty
}

func parseVerification(doc *openapi3.T, refName string) (*Verification, error) {
	log.Println("here", refName)
	// Get verification data
	verificationSchemaAny, err := doc.Components.Schemas.JSONLookup(refName)
	if err != nil {
		return nil, err
	}
	verificationSchema, ok := verificationSchemaAny.(*openapi3.Schema)
	if !ok {
		return nil, errors.New("type assertion failed")
	}
	// verificationName, ok := verificationSchema.Properties["type"].Value.Enum[0].(string)
	verificationName := ""
	if verificationSchema.Properties["type"] != nil {
		verificationName, ok = verificationSchema.Properties["type"].Value.Enum[0].(string)
		if !ok {
			return nil, errors.New("type assertion failed")
		}
	}
	verificationName = strings.ReplaceAll(verificationName, "-", "_")
	docsType, ok := verificationSchema.Extensions["x-docs-type"].(string)
	if !ok {
		return nil, errors.New("type assertion failed")
	}

	log.Println("name", verificationName, "docsType", docsType)
	verificationName = docsType

	// Get verification config data
	// verificationConfigSchemaRef := verificationSchema.Properties["configs"].Ref
	// verificationConfigSchemaRefName := getSchemaNameFromRef(verificationConfigSchemaRef)

	// verificationConfigSchemaAny, err := doc.Components.Schemas.JSONLookup(verificationConfigSchemaRefName)
	// if err != nil {
	// 	return nil, err
	// }
	// verificationConfigSchema, ok := verificationConfigSchemaAny.(*openapi3.Schema)
	// if !ok {
	// 	return nil, errors.New("type assertion failed")
	// }

	// log.Println("here", verificationName, verificationConfigSchema)

	// Constructing Verification struct

	verificationConfigSchemaRefName := "TODO"
	verificationConfigSchema := verificationSchema.Properties["auth"].Value

	properties := []VerificationProperty{}

	// for _, key := range getProperties(verificationName, verificationConfigSchema) {
	for _, key := range getProperties(verificationName, verificationConfigSchema) {
		property := verificationConfigSchema.Properties[key]
		log.Println("key", key, property)
		required := slices.Contains(verificationConfigSchema.Required, key)
		pointerString := ""
		if !required {
			pointerString = "Pointer"
		}
		typeString := ""
		if property.Value.Type.Is("string") {
			typeString = "String"
		} else if property.Value.Type.Is("number") && property.Value.Format == "float" {
			typeString = "Float64"
		}
		isEnum := false
		enumNameString := ""
		if len(property.Value.Enum) > 0 {
			isEnum = true
			if property.Ref == "" {
				enumNameString = "New" + toConfigCase(verificationConfigSchemaRefName) + toConfigCase(key) + "From" + typeString
			} else {
				enumNameString = "New" + toConfigCase(getSchemaNameFromRef(property.Ref)) + "From" + typeString
			}
		}

		properties = append(properties, VerificationProperty{
			NameSnake:      key,
			NameCamel:      strcase.ToLowerCamel(key),
			NamePascal:     strcase.ToCamel(key),
			Optional:       !required,
			Required:       required,
			Sensitive:      true,
			PointerString:  pointerString,
			TypeString:     typeString,
			IsEnum:         isEnum,
			EnumNameString: enumNameString,
		})
	}

	verification := &Verification{
		Ref:        refName,
		ConfigRef:  verificationConfigSchemaRefName,
		NameSnake:  strcase.ToSnake(verificationName),
		NameCamel:  strcase.ToLowerCamel(verificationName),
		NamePascal: strcase.ToCamel(verificationName),
		NameConfig: toConfigCase(docsType),
		Properties: properties,
	}

	return verification, nil
}

func getProperties(verificationName string, schema *openapi3.Schema) []string {
	if verificationName == "shopify" {
		// edge case: for Shopify we only care about "webhook_secret_key"
		return []string{"webhook_secret_key"}
	}

	properties := []string{}

	for key := range schema.Properties {
		properties = append(properties, key)
	}

	sort.Strings(properties)

	return properties
}
