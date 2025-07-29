package connection

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func schemaAttributes() map[string]schema.Attribute {
	return schemaAttributesV1()
}
