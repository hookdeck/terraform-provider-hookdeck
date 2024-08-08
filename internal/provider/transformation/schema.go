package transformation

import (
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"code": schema.StringAttribute{
			Required:    true,
			Description: "JavaScript code to be executed",
		},
		"created_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "Date the transformation was created",
		},
		"env": schema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "Key-value environment variables to be passed to the transformation",
		},
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the transformation",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique, human-friendly name for the transformation",
		},
		"team_id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the workspace",
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			Description: "Date the transformation was last updated",
		},
	}
}
