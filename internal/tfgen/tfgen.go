package tfgen

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-codegen-spec/provider"
	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
	"github.com/hashicorp/terraform-plugin-codegen-spec/schema"
	"github.com/hashicorp/terraform-plugin-codegen-spec/spec"
)

func Generate() error {
	doc, err := loadOpenAPISchema()
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}
	log.Println(doc.Info.Title)

	provider := provider.Provider{
		Name: "hookdeck",
	}
	sourceResource := resource.Resource{
		Name: "source",
		Schema: &resource.Schema{
			Attributes: []resource.Attribute{
				{
					Name: "id",
					String: &resource.StringAttribute{
						ComputedOptionalRequired: schema.Computed,
					},
				},
			},
		},
	}
	spec := spec.Specification{
		Version:  spec.Version0_1,
		Provider: &provider,
		Resources: []resource.Resource{
			sourceResource,
		},
	}

	bytes, err := json.MarshalIndent(spec, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshalling provider code spec to JSON: %w", err)
	}

	if err := writeTFCodeSpec(bytes); err != nil {
		return fmt.Errorf("failed to write TF code spec: %w", err)
	}

	return nil
}
