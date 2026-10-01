package project

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// The API trims the name; a padded value would come back different from
// the plan.
var noSurroundingSpace = regexp.MustCompile(`^\S(.*\S)?$`)

func schemaAttributes(k kind) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"created_at": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "Date the project was created",
		},
		"id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "ID of the project",
		},
		"name": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 40),
				stringvalidator.RegexMatches(noSurroundingSpace, "must not start or end with whitespace"),
			},
			Description: "Name of the project",
		},
		"organization_id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   "ID of the organization the project belongs to",
		},
		"type": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			Description:   fmt.Sprintf("Project type, always `%s`", k.apiType),
		},
		"updated_at": schema.StringAttribute{
			Computed:    true,
			Description: "Date the project was last updated",
		},
	}
}
