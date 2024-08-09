package sourceverification

import (
	"terraform-provider-hookdeck/internal/provider/sourceverification/generated"
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func schemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source_id": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Description: "ID of the source",
		},
		"verification": schema.SingleNestedAttribute{
			Required:   true,
			Attributes: generated.GetSourceVerificationSchemaAttributes(),
			Validators: []validator.Object{
				validators.ExactlyOneChild(),
			},
			Description: "The verification configs for the specific verification type",
		},
	}
}
