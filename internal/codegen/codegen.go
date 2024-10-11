package codegen

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const hookdeckOpenAPISchemaURI = "https://api.hookdeck.com/latest/openapi"

func RunCodeGen() error {
	fmt.Println("generating Hookdeck source verifications")

	// Load OpenAPI schema
	loader := openapi3.NewLoader()
	u, err := url.Parse(hookdeckOpenAPISchemaURI)
	if err != nil {
		return err
	}
	fmt.Println("getting Hookdeck OpenAPI schema")
	doc, err := loader.LoadFromURI(u)
	if err != nil {
		return err
	}

	// Get list of source verification providers
	verificationListAny, err := doc.Components.Schemas.JSONLookup("VerificationConfig")
	if err != nil {
		return err
	}
	verificationList, ok := verificationListAny.(*openapi3.Schema)
	if !ok {
		return errors.New("type assertion failed")
	}

	// Construct verification data
	verifications := []Verification{}
	for _, schemaRef := range verificationList.OneOf {
		verification, err := parseVerification(doc, getSchemaNameFromRef(schemaRef.Ref))
		if err != nil {
			return err
		}
		verifications = append(verifications, *verification)
	}

	sort.Slice(verifications, func(i, j int) bool {
		return verifications[i].NamePascal < verifications[j].NamePascal
	})

	// Generate code using template
	if err := generateModel(verifications); err != nil {
		return err
	}
	if err := generateVerifications(verifications); err != nil {
		return err
	}

	return nil
}

func getSchemaNameFromRef(ref string) string {
	// Split the string by '/' and take the last element
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
