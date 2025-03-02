package destination

import (
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"config": schema.StringAttribute{
			Optional: true,
			// Cannot be computed because some destinations may have default config value,
			// leading a conflict between the initial state & computed state after creation.
			Computed:    false,
			Description: "Destination configuration",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			CustomType: jsontypes.NormalizedType{},
		},
		"created_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: `Date the destination was created`,
		},
		"description": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "Description for the destination",
		},
		"disabled_at": schema.StringAttribute{
			Computed: true,
			Optional: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			Description: `Date the destination was disabled`,
		},
		"id": schema.StringAttribute{
			Computed:    true,
			Description: `ID of the destination`,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: `A unique, human-friendly name for the destination`,
		},
		"team_id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: `ID of the workspace`,
		},
		"type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Type of the destination",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			Description: `Date the destination was last updated`,
		},
	}
}
