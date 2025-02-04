package tfgen

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-codegen-spec/code"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
)

func Generate() error {
	doc, err := loadOpenAPISchema()
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}

	sourceTypeNames, err := getSourceTypes(doc)
	if err != nil {
		return fmt.Errorf("failed to get source types: %w", err)
	}

	sourceConfigAttributes := []resource.Attribute{}
	resources := []resource.Resource{}
	sourceTypeNames = []string{"SourceTypeConfigEbay", "SourceTypeConfigHTTP", "SourceTypeConfigShopify"}
	for _, sourceTypeName := range sourceTypeNames {
		fmt.Println(sourceTypeName) // TODO: remove
		authResource, sourceConfigAttribute := parseSourceConfig(doc, sourceTypeName)
		resources = append(resources, authResource)
		sourceConfigAttributes = append(sourceConfigAttributes, sourceConfigAttribute)
	}

	sourceAttributes := parseSourceAttributes(doc)
	sourceAttributes = append(sourceAttributes, resource.Attribute{
		Name: "config",
		SingleNested: &resource.SingleNestedAttribute{
			Attributes:               sourceConfigAttributes,
			ComputedOptionalRequired: schema.Required,
			Validators: []schema.ObjectValidator{
				{
					Custom: &schema.CustomValidator{
						Imports: []code.Import{
							{
								Path: "terraform-provider-hookdeck/internal/validators",
							},
						},
						SchemaDefinition: "validators.ExactlyOneChild()",
					},
				},
			},
		},
	})
	resources = append(resources, resource.Resource{
		Name: "source",
		Schema: &resource.Schema{
			Attributes: sourceAttributes,
		},
	})
	spec := makeSpec(resources)

	if err := writeTFCodeSpec(spec); err != nil {
		return fmt.Errorf("failed to write TF code spec: %w", err)
	}

	return nil
}
