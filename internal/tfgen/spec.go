package tfgen

import (
	"fmt"
	"slices"
	"sort"
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

func parseSourceAttributes(doc *openapi3.T) []resource.Attribute {
	attributes := []resource.Attribute{}

	attributes = append(attributes, resource.Attribute{
		Name: "id",
		String: &resource.StringAttribute{
			ComputedOptionalRequired: schema.Computed,
		},
	})

	sourceSchemaAny, err := doc.Components.Schemas.JSONLookup("Source")
	if err != nil {
		panic(err)
	}
	sourceSchema, ok := sourceSchemaAny.(*openapi3.Schema)
	if !ok {
		panic(err)
	}

	includedFields := []struct {
		fieldName                string
		computedOptionalRequired schema.ComputedOptionalRequired
	}{
		{"name", schema.Required},
		{"description", schema.Optional},
		{"team_id", schema.Computed},
		{"url", schema.Computed},
		{"disabled_at", schema.Computed},
		{"updated_at", schema.Computed},
		{"created_at", schema.Computed},
	}
	includedFieldMap := make(map[string]struct {
		fieldName                string
		computedOptionalRequired schema.ComputedOptionalRequired
	})
	for _, field := range includedFields {
		includedFieldMap[field.fieldName] = struct {
			fieldName                string
			computedOptionalRequired schema.ComputedOptionalRequired
		}{
			fieldName:                field.fieldName,
			computedOptionalRequired: field.computedOptionalRequired,
		}
	}

	for fieldName, field := range sourceSchema.Properties {
		includedField, ok := includedFieldMap[fieldName]
		if !ok {
			continue
		}
		attr := parseSchemaField(doc, fieldName, field.Value, includedField.computedOptionalRequired)
		attributes = append(attributes, attr)
	}

	return attributes
}

func parseSourceConfig(doc *openapi3.T, sourceTypeName string) (resource.Resource, resource.Attribute) {
	name := strings.TrimPrefix(sourceTypeName, "SourceTypeConfig")
	name = strings.ReplaceAll(name, ".", "")
	nameSnake := strcase.ToSnake(name)

	sourceAttributes, authAttributes := parseSourceConfigAttributes(doc, sourceTypeName)

	authAttributes = append(authAttributes, resource.Attribute{
		Name: "source_id",
		String: &resource.StringAttribute{
			ComputedOptionalRequired: schema.Required,
		},
	})

	authResource := resource.Resource{
		Name: fmt.Sprintf("source_config_%s", nameSnake),
		Schema: &resource.Schema{
			Attributes: authAttributes,
		},
	}

	return authResource, resource.Attribute{
		Name: nameSnake,
		SingleNested: &resource.SingleNestedAttribute{
			Attributes:               sourceAttributes,
			ComputedOptionalRequired: schema.Optional,
		},
	}
}

func parseSourceConfigAttributes(doc *openapi3.T, sourceTypeName string) ([]resource.Attribute, []resource.Attribute) {
	sourceAttributes := []resource.Attribute{}
	authAttributes := []resource.Attribute{}

	sourceSchemaAny, err := doc.Components.Schemas.JSONLookup(sourceTypeName)
	if err != nil {
		panic(err)
	}
	sourceSchema, ok := sourceSchemaAny.(*openapi3.Schema)
	if !ok {
		panic(err)
	}

	for fieldName, field := range sourceSchema.Properties {
		if strings.Contains(fieldName, "auth") {
			if field.Value.OneOf != nil {
				attributes := parseAuthOneOf(doc, fieldName, field, getComputedOptionalRequired(sourceSchema.Required, fieldName))
				authAttributes = append(authAttributes, attributes...)
			} else if field.Value == nil {
				ref, err := doc.Components.Schemas.JSONLookup(getNameFromRef(field.Ref))
				if err != nil {
					fmt.Println(fieldName)
					fmt.Printf("error looking up ref: %v", err)
					panic(err)
				}
				refSchema, ok := ref.(*openapi3.Schema)
				if !ok {
					panic(fmt.Sprintf("expected schema, got %T", ref))
				}
				attr := parseSchemaField(doc, fieldName, refSchema, getComputedOptionalRequired(sourceSchema.Required, fieldName))
				authAttributes = append(authAttributes, attr)
			} else {
				attr := parseSchemaField(doc, fieldName, field.Value, getComputedOptionalRequired(sourceSchema.Required, fieldName))
				authAttributes = append(authAttributes, attr)
			}
		} else if fieldName == "type" {
			attr := parseSchemaField(doc, fieldName, field.Value, getComputedOptionalRequired(sourceSchema.Required, fieldName))
			attr.Name = "auth_type"
			authAttributes = append(authAttributes, attr)
		} else {
			attr := parseSchemaField(doc, fieldName, field.Value, getComputedOptionalRequired(sourceSchema.Required, fieldName))
			sourceAttributes = append(sourceAttributes, attr)
		}
	}

	return sourceAttributes, authAttributes
}

func getComputedOptionalRequired(required []string, fieldName string) schema.ComputedOptionalRequired {
	if slices.Contains(required, fieldName) {
		return schema.Required
	}
	return schema.ComputedOptional
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

func parseSchemaField(doc *openapi3.T, fieldName string, field *openapi3.Schema, computedOptionalRequired schema.ComputedOptionalRequired) resource.Attribute {
	attr := resource.Attribute{Name: fieldName}

	switch {
	case field.Type.Is("string"):
		stringAttr := &resource.StringAttribute{}
		stringAttr.ComputedOptionalRequired = computedOptionalRequired
		if computedOptionalRequired != schema.Required {
			stringAttr.PlanModifiers = []schema.StringPlanModifier{
				{
					Custom: &schema.CustomPlanModifier{
						Imports: []code.Import{
							{
								Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier",
							},
						},
						SchemaDefinition: "stringplanmodifier.UseStateForUnknown()",
					},
				},
			}
		}

		if len(field.Enum) > 0 {
			enumVals := make([]string, 0, len(field.Enum))
			for _, enum := range field.Enum {
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
		attr.String = stringAttr
	case field.Type.Is("array"):
		if items := field.Items; items != nil {
			if items.Value.Type.Is("object") {
				nestedAttrs := []resource.Attribute{}
				for propName, prop := range items.Value.Properties {
					nestedAttr := parseSchemaField(doc, propName, prop.Value, getComputedOptionalRequired(items.Value.Required, propName))
					nestedAttrs = append(nestedAttrs, nestedAttr)
				}
				attr.ListNested = &resource.ListNestedAttribute{
					NestedObject: resource.NestedAttributeObject{
						Attributes: nestedAttrs,
					},
					ComputedOptionalRequired: computedOptionalRequired,
				}
			} else {
				listAttr := &resource.ListAttribute{
					ComputedOptionalRequired: computedOptionalRequired,
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

				attr.List = listAttr
			}
		}
	case field.Type.Is("object"):
		nestedAttrs := []resource.Attribute{}
		for propName, prop := range field.Properties {
			nestedAttr := parseSchemaField(doc, propName, prop.Value, getComputedOptionalRequired(field.Required, propName))
			nestedAttrs = append(nestedAttrs, nestedAttr)
		}
		attr.SingleNested = &resource.SingleNestedAttribute{
			Attributes:               nestedAttrs,
			ComputedOptionalRequired: computedOptionalRequired,
		}
		if computedOptionalRequired != schema.Required {
			attr.SingleNested.PlanModifiers = []schema.ObjectPlanModifier{
				{
					Custom: &schema.CustomPlanModifier{
						Imports: []code.Import{
							{
								Path: "github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier",
							},
						},
						SchemaDefinition: "objectplanmodifier.UseStateForUnknown()",
					},
				},
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
				SchemaDefinition: "listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue(\"POST\"), types.StringValue(\"PUT\"), types.StringValue(\"PATCH\"), types.StringValue(\"DELETE\")}))",
			},
		}
	}

	return attr
}

func parseAuthOneOf(doc *openapi3.T, fieldName string, field *openapi3.SchemaRef, computedOptionalRequired schema.ComputedOptionalRequired) []resource.Attribute {
	attributes := []resource.Attribute{}

	for _, oneOf := range field.Value.OneOf {
		attr := parseSchemaField(doc, fieldName, oneOf.Value, computedOptionalRequired)
		attr.Name = getAuthOneOfName(attr)
		attributes = append(attributes, attr)
	}

	return attributes
}

func getAuthOneOfName(attr resource.Attribute) string {
	attributeNames := []string{}
	for _, a := range attr.SingleNested.Attributes {
		attributeNames = append(attributeNames, a.Name)
	}
	sort.Strings(attributeNames)
	switch attributeNames[0] {
	case "algorithm":
		return "auth_hmac"
	case "password":
		return "auth_basic"
	case "api_key":
		return "auth_api_key"
	default:
		panic(fmt.Sprintf("unknown one of: %v", attributeNames))
	}
}

func getNameFromRef(ref string) string {
	return strings.TrimPrefix(ref, "#/components/schemas/")
}
