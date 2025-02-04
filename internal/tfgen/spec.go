package tfgen

import (
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/hashicorp/terraform-plugin-codegen-spec/code"
	"github.com/hashicorp/terraform-plugin-codegen-spec/provider"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
	"github.com/hashicorp/terraform-plugin-codegen-spec/spec"
	"github.com/iancoleman/strcase"
)

func makeSpec(resources []resource.Resource) spec.Specification {
	provider := provider.Provider{
		Name: "hookdeck",
	}
	spec := spec.Specification{
		Version:   spec.Version0_1,
		Provider:  &provider,
		Resources: resources,
	}

	return spec
}

func parseSourceConfig(doc *openapi3.T, sourceTypeName string) resource.Resource {
	name := strings.TrimPrefix(sourceTypeName, "SourceTypeConfig")
	name = strings.ReplaceAll(name, ".", "")
	nameSnake := strcase.ToSnake(name)

	attributes := parseSourceConfigAttributes(doc, sourceTypeName)

	attributes = append(attributes, resource.Attribute{
		Name: "source_id",
		String: &resource.StringAttribute{
			ComputedOptionalRequired: schema.Required,
		},
	})

	sourceConfigResource := resource.Resource{
		Name: fmt.Sprintf("source_config_%s", nameSnake),
		Schema: &resource.Schema{
			Attributes: attributes,
		},
	}

	return sourceConfigResource
}

func parseSourceConfigAttributes(doc *openapi3.T, sourceTypeName string) []resource.Attribute {
	attributes := []resource.Attribute{}

	sourceSchemaAny, err := doc.Components.Schemas.JSONLookup(sourceTypeName)
	if err != nil {
		return attributes
	}
	sourceSchema, ok := sourceSchemaAny.(*openapi3.Schema)
	if !ok {
		return attributes
	}

	for fieldName, field := range sourceSchema.Properties {
		log.Println("fieldName", fieldName)
		attr := parseSchemaField(fieldName, field, sourceSchema.Required)
		attributes = append(attributes, attr)
	}

	return attributes
}

func getElementType(field *openapi3.SchemaRef) schema.ElementType {
	elementType := schema.ElementType{}
	if field.Value == nil {
		return elementType
	}

	switch {
	case field.Value.Type.Is("string"):
		elementType.String = &schema.StringType{}
	case field.Value.Type.Is("number"):
		elementType.Number = &schema.NumberType{}
	case field.Value.Type.Is("integer"):
		elementType.Int64 = &schema.Int64Type{}
	case field.Value.Type.Is("boolean"):
		elementType.Bool = &schema.BoolType{}
	}

	return elementType
}

func parseSchemaField(fieldName string, field *openapi3.SchemaRef, required []string) resource.Attribute {
	isRequired := false
	for _, req := range required {
		if req == fieldName {
			isRequired = true
			break
		}
	}

	attr := resource.Attribute{Name: fieldName}

	if field.Value != nil {
		switch {
		case field.Value.Type.Is("string"):
			stringAttr := &resource.StringAttribute{
				ComputedOptionalRequired: schema.ComputedOptional,
			}
			if len(field.Value.Enum) > 0 {
				enumVals := make([]string, 0, len(field.Value.Enum))
				for _, enum := range field.Value.Enum {
					if str, ok := enum.(string); ok {
						enumVals = append(enumVals, str)
					}
				}
				if len(enumVals) > 0 {
					var enumDef strings.Builder
					enumDef.WriteString("stringvalidator.OneOf(\n")
					for _, val := range enumVals {
						enumDef.WriteString(fmt.Sprintf("%q,\n", val))
					}
					enumDef.WriteString(")")

					stringAttr.Validators = []schema.StringValidator{
						{
							Custom: &schema.CustomValidator{
								Imports: []code.Import{
									{
										Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
									},
								},
								SchemaDefinition: enumDef.String(),
							},
						},
					}
				}
			}
			if isRequired {
				stringAttr.ComputedOptionalRequired = schema.Required
			}
			attr.String = stringAttr
		case field.Value.Type.Is("array"):
			if items := field.Value.Items; items != nil {
				if items.Value.Type.Is("object") {
					nestedAttrs := []resource.Attribute{}
					for propName, prop := range items.Value.Properties {
						nestedAttr := parseSchemaField(propName, prop, items.Value.Required)
						nestedAttrs = append(nestedAttrs, nestedAttr)
					}
					attr.ListNested = &resource.ListNestedAttribute{
						NestedObject: resource.NestedAttributeObject{
							Attributes: nestedAttrs,
						},
						ComputedOptionalRequired: schema.ComputedOptional,
					}
					if isRequired {
						attr.ListNested.ComputedOptionalRequired = schema.Required
					}
				} else {
					listAttr := &resource.ListAttribute{
						ComputedOptionalRequired: schema.ComputedOptional,
						ElementType:              getElementType(items),
					}

					// Add enum validation for string arrays
					if items.Value.Type.Is("string") && len(items.Value.Enum) > 0 {
						enumVals := make([]string, 0, len(items.Value.Enum))
						for _, enum := range items.Value.Enum {
							if str, ok := enum.(string); ok {
								enumVals = append(enumVals, str)
							}
						}
						if len(enumVals) > 0 {
							var enumDef strings.Builder
							enumDef.WriteString("listvalidator.ValueStringsAre(\n")
							enumDef.WriteString("\tstringvalidator.OneOf(\n")
							for _, val := range enumVals {
								enumDef.WriteString(fmt.Sprintf("\t\t%q,\n", val))
							}
							enumDef.WriteString("\t),\n")
							enumDef.WriteString(")")

							listAttr.Validators = []schema.ListValidator{
								{
									Custom: &schema.CustomValidator{
										Imports: []code.Import{
											{
												Path: "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator",
											},
											{
												Path: "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator",
											},
										},
										SchemaDefinition: enumDef.String(),
									},
								},
							}
						}
					}

					if isRequired {
						listAttr.ComputedOptionalRequired = schema.Required
					}
					attr.List = listAttr
				}
			}
		case field.Value.Type.Is("object"):
			nestedAttrs := []resource.Attribute{}
			for propName, prop := range field.Value.Properties {
				nestedAttr := parseSchemaField(propName, prop, field.Value.Required)
				nestedAttrs = append(nestedAttrs, nestedAttr)
			}
			attr.SingleNested = &resource.SingleNestedAttribute{
				Attributes:               nestedAttrs,
				ComputedOptionalRequired: schema.ComputedOptional,
			}
			if isRequired {
				attr.SingleNested.ComputedOptionalRequired = schema.Required
			}
		}
	}

	if fieldName == "allowed_http_methods" {
		attr.List.Default = &schema.ListDefault{
			Custom: &schema.CustomDefault{
				Imports: []code.Import{
					{
						Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault",
					},
				},
				SchemaDefinition: "listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue(\"POST\"), types.StringValue(\"PUT\"), types.StringValue(\"PATCH\"), types.StringValue(\"GET\")}))",
			},
		}
	}

	return attr
}
