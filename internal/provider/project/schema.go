package project

import (
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const projectType = "event_gateway"

func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"created_at": schema.StringAttribute{
			Computed:      true,
			Validators:    []validator.String{validators.IsRFC3339()},
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "Date the project was created",
		},
		"headers_prefix": schema.StringAttribute{
			Optional:    true,
			Description: "Prefix for the headers Hookdeck adds to delivered requests",
		},
		"id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "ID of the project",
		},
		"max_events_per_second": schema.Int64Attribute{
			Computed:      true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			Description:   "Maximum events per second for the project",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Validators:  []validator.String{stringvalidator.LengthBetween(1, 40)},
			Description: "Name of the project",
		},
		"notification_methods": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf("email", "webhook")),
			},
			PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			Description:   "Notification methods enabled for the project: `email`, `webhook`",
		},
		"organization_id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "ID of the organization the project belongs to",
		},
		"type": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "Project type, always `event_gateway`",
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Validators:  []validator.String{validators.IsRFC3339()},
			Description: "Date the project was last updated",
		},
	}
}
