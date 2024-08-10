package codegen

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const hookdeckOpenAPISchemaURI = "https://api.hookdeck.com/2024-03-01/openapi"

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
	verificationList := verificationListAny.(*openapi3.Schema)

	// Construct verification data
	verifications := []Verification{}
	for _, schemaRef := range verificationList.OneOf {
		// SDK not supported yet
		if schemaRef.Ref == "#/components/schemas/VerificationMailchimp" || schemaRef.Ref == "#/components/schemas/VerificationPaddle" {
			continue
		}

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
	generateModel(verifications)
	generateVerifications(verifications)

	return nil
}

func getSchemaNameFromRef(ref string) string {
	// Split the string by '/' and take the last element
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
