package codegenv2

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// const hookdeckOpenAPISchemaURI = "https://raw.githubusercontent.com/hookdeck/hookdeck-api-schema/refs/heads/main/openapi.json"
const hookdeckOpenAPISchemaURI = "internal/codegenv2/openapi.json"

func RunCodeGen() error {
	fmt.Println("generating Hookdeck sources")

	fmt.Println("getting Hookdeck OpenAPI schema")
	doc, err := loadOpenAPISchema(hookdeckOpenAPISchemaURI)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}

	sourceTypeNames, err := getSourceTypes(doc)
	if err != nil {
		return fmt.Errorf("failed to get source types: %w", err)
	}

	// TODO: remove
	// sourceTypeNames = []string{"SourceTypeConfigShopify"}

	// Construct source type data
	sourceTypes := []SourceType{}
	for _, sourceTypeName := range sourceTypeNames {
		sourceType, err := parseSourceType(doc, sourceTypeName)
		if err != nil {
			log.Println("error parsing source", sourceTypeName, err)
			// return err
			continue
		}
		sourceTypes = append(sourceTypes, *sourceType)
		log.Println(sourceType)
	}

	sort.Slice(sourceTypes, func(i, j int) bool {
		return sourceTypes[i].NamePascal < sourceTypes[j].NamePascal
	})

	// TODO: Generate source files
	if err := generateSources(sourceTypes); err != nil {
		return fmt.Errorf("failed to generate sources: %w", err)
	}

	return nil
}

func loadOpenAPISchema(source string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		u, err := url.Parse(source)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URI: %w", err)
		}
		return loader.LoadFromURI(u)
	}

	return loader.LoadFromFile(source)
}

func getSourceTypes(doc *openapi3.T) ([]string, error) {
	// Get list of source verification providers
	sourceTypesAny, err := doc.Components.Schemas.JSONLookup("SourceTypeConfig")
	if err != nil {
		return nil, err
	}
	sourceTypes, ok := sourceTypesAny.(*openapi3.Schema)
	if !ok {
		return nil, errors.New("type assertion failed")
	}

	sourceTypeNames := []string{}
	for _, schemaRef := range sourceTypes.OneOf {
		sourceTypeNames = append(sourceTypeNames, getSchemaNameFromRef(schemaRef.Ref))
	}

	return sourceTypeNames, nil
}

func getSchemaNameFromRef(ref string) string {
	// Split the string by '/' and take the last element
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
