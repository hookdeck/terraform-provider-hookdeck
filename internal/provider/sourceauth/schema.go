package sourceauth

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"auth": schema.StringAttribute{
			Required:    true,
			Description: "Source auth",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			CustomType: jsontypes.NormalizedType{},
		},
		"auth_type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Type of the source auth",
		},
		"source_id": schema.StringAttribute{
			Required:    true,
			Description: "ID of the source",
		},
	}
}
