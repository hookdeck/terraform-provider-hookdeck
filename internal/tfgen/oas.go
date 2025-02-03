package tfgen

import (
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
