package tfgen

import (
	"fmt"
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
		attr := parseSchemaField(fieldName, field, sourceSchema.Required)
		attributes = append(attributes, attr)
	}

	return attributes
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
				ComputedOptionalRequired: schema.Optional,
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
				attr.List = &resource.ListAttribute{
					ComputedOptionalRequired: schema.Optional,
					ElementType: schema.ElementType{
						String: &schema.StringType{},
					},
				}
				if isRequired {
					attr.List.ComputedOptionalRequired = schema.Required
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
				ComputedOptionalRequired: schema.Optional,
			}
			if isRequired {
				attr.SingleNested.ComputedOptionalRequired = schema.Required
			}
		}
	}

	return attr
}
