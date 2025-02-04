package tfgen

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-codegen-spec/resource"
)

func Generate() error {
	doc, err := loadOpenAPISchema()
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}
	log.Println(doc.Info.Title)

	sourceTypeNames, err := getSourceTypes(doc)
	if err != nil {
		return fmt.Errorf("failed to get source types: %w", err)
	}

	resources := []resource.Resource{}
	sourceTypeNames = []string{"SourceTypeConfigEbay", "SourceTypeConfigWebhook"}
	for _, sourceTypeName := range sourceTypeNames {
		log.Println(sourceTypeName)
		sourceResource := parseSourceConfig(doc, sourceTypeName)
		resources = append(resources, sourceResource)
	}

	spec := makeSpec(resources)

	if err := writeTFCodeSpec(spec); err != nil {
		return fmt.Errorf("failed to write TF code spec: %w", err)
	}

	return nil
}
