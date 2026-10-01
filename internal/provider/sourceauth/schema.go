package sourceauth

import (
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func schemaAttributes() map[string]schema.Attribute {
	projectID := shared.ProjectIDResourceAttribute()
	projectID.Description = "ID of the project the source belongs to. With a project API key or a provider `project_id`, it can be omitted. With an organization API key and no provider `project_id`, it is required."
	return map[string]schema.Attribute{
		"project_id": projectID,
		"auth": schema.StringAttribute{
			Required:    true,
			Sensitive:   true,
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
