package codegenv2

import (
	"errors"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

func parseSourceType(doc *openapi3.T, refName string) (*SourceType, error) {
	// Get source type data
	sourceSchemaAny, err := doc.Components.Schemas.JSONLookup(refName)
	if err != nil {
		return nil, err
	}
	sourceSchema, ok := sourceSchemaAny.(*openapi3.Schema)
	if !ok {
		return nil, errors.New("type assertion failed")
	}

	// Extract and clean name from schema
	name := strings.TrimPrefix(refName, "SourceTypeConfig")
	name = strings.ReplaceAll(name, ".", "")
	typeEnum := strings.ToUpper(name)
	nameCamel := strcase.ToLowerCamel(name)
	nameSnake := strcase.ToSnake(name)

	// Parse auth fields
	authFields := []AuthField{}
	if authSchema := sourceSchema.Properties["auth"]; authSchema != nil && authSchema.Value != nil {
		for fieldName, field := range authSchema.Value.Properties {
			// Check if field is required
			required := false
			for _, req := range authSchema.Value.Required {
				if req == fieldName {
					required = true
					break
				}
			}

			// Check if field is nullable
			nullable := false
			if field.Value != nil {
				nullable = field.Value.Nullable
			}

			// Check if field is enum and get enum values
			isEnum := false
			enumNameString := ""
			enumValues := []string{}
			if field.Value != nil && len(field.Value.Enum) > 0 {
				isEnum = true
				if field.Ref == "" {
					enumNameString = "hookdeck.New" + toConfigCase(refName) + "Auth" + toConfigCase(fieldName) + "FromString"
				} else {
					enumNameString = "hookdeck.New" + toConfigCase(getSchemaNameFromRef(field.Ref)) + "FromString"
				}
				for _, enum := range field.Value.Enum {
					if str, ok := enum.(string); ok {
						enumValues = append(enumValues, str)
					}
				}
			}

			authFields = append(authFields, AuthField{
				NameSnake:      fieldName,
				NameCamel:      strcase.ToLowerCamel(fieldName),
				NamePascal:     strcase.ToCamel(fieldName),
				Required:       required,
				Nullable:       nullable,
				IsEnum:         isEnum,
				EnumNameString: enumNameString,
				EnumValues:     enumValues,
			})
		}
	}

	return &SourceType{
		Ref:        refName,
		ConfigRef:  refName,
		NameSnake:  nameSnake,
		NameCamel:  nameCamel,
		NamePascal: name,
		NameConfig: toConfigCase(name),
		TypeEnum:   typeEnum,
		Auth:       authFields,
	}, nil
}

func getProperties(sourceType string, schema *openapi3.Schema) []string {
	properties := []string{}

	for key := range schema.Properties {
		properties = append(properties, key)
	}

	sort.Strings(properties)

	return properties
}
