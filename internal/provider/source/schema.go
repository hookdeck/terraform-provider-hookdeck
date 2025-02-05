package source

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
			// Cannot be computed because some sources may have default config value,
			// leading a conflict between the initial state & computed state after creation.
			Computed:    false,
			Description: "Source configuration",
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
			Description: "Date the source was created",
		},
		"description": schema.StringAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "Description for the source",
		},
		"disabled_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			Description: "Date the source was disabled",
		},
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the source",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique, human-friendly name for the source",
		},
		"team_id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the workspace",
		},
		"type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Type of the source",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
			Validators: []validator.String{
				validators.IsRFC3339(),
			},
			Description: "Date the source was last updated",
		},
		"url": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "A unique URL that must be supplied to your webhook's provider",
		},
	}
}
