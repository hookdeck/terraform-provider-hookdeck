package tfgen

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// const hookdeckOpenAPISchemaURI = "https://raw.githubusercontent.com/hookdeck/hookdeck-api-schema/refs/heads/main/openapi.json"
const hookdeckOpenAPISchemaURI = "internal/tfgen/openapi.json"

func loadOpenAPISchema() (*openapi3.T, error) {
	loader := openapi3.NewLoader()

	if strings.HasPrefix(hookdeckOpenAPISchemaURI, "http://") || strings.HasPrefix(hookdeckOpenAPISchemaURI, "https://") {
		u, err := url.Parse(hookdeckOpenAPISchemaURI)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URI: %w", err)
		}
		return loader.LoadFromURI(u)
	}

	return loader.LoadFromFile(hookdeckOpenAPISchemaURI)
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
