package codegenv2

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/iancoleman/strcase"
)

func parseField(fieldName string, field *openapi3.SchemaRef, required []string) Field {
	isRequired := false
	for _, req := range required {
		if req == fieldName {
			isRequired = true
			break
		}
	}

	var fieldType FieldType
	if field.Value != nil {
		switch {
		case field.Value.Type.Is("string"):
			enumValues := []string{}
			for _, enum := range field.Value.Enum {
				if str, ok := enum.(string); ok {
					enumValues = append(enumValues, str)
				}
			}
			fieldType = StringField{
				IsEnum:     len(enumValues) > 0,
				EnumValues: enumValues,
			}
			fmt.Printf("Parsed string field %s: %+v\n", fieldName, fieldType)
		case field.Value.Type.Is("array"):
			if items := field.Value.Items; items != nil {
				fieldType = ArrayField{
					ItemType: parseField("", items, nil).Type,
				}
				fmt.Printf("Parsed array field %s: %+v\n", fieldName, fieldType)
			}
		case field.Value.Type.Is("object"):
			objField := ObjectField{
				Properties: []Field{},
				Required:   field.Value.Required,
			}
			if len(field.Value.OneOf) > 0 {
				for _, oneOf := range field.Value.OneOf {
					if oneOf.Value != nil {
						oneOfObj := ObjectField{
							Properties: []Field{},
							Required:   oneOf.Value.Required,
						}
						for propName, prop := range oneOf.Value.Properties {
							oneOfObj.Properties = append(oneOfObj.Properties, parseField(propName, prop, oneOf.Value.Required))
						}
						objField.OneOf = append(objField.OneOf, oneOfObj)
					}
				}
			} else {
				for propName, prop := range field.Value.Properties {
					objField.Properties = append(objField.Properties, parseField(propName, prop, field.Value.Required))
				}
			}
			fieldType = objField
			fmt.Printf("Parsed object field %s: %+v\n", fieldName, fieldType)
		}
	}

	return Field{
		NameSnake:  fieldName,
		NameCamel:  strcase.ToLowerCamel(fieldName),
		NamePascal: strcase.ToCamel(fieldName),
		Required:   isRequired,
		Nullable:   field.Value != nil && field.Value.Nullable,
		Type:       fieldType,
	}
}

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

	// Parse all fields
	fields := []Field{}
	for fieldName, field := range sourceSchema.Properties {
		fields = append(fields, parseField(fieldName, field, sourceSchema.Required))
	}

	return &SourceType{
		Ref:        refName,
		ConfigRef:  refName,
		NameSnake:  nameSnake,
		NameCamel:  nameCamel,
		NamePascal: name,
		NameConfig: toConfigCase(name),
		TypeEnum:   typeEnum,
		Fields:     fields,
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
