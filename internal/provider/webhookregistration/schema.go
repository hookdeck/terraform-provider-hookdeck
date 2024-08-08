package webhookregistration

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func schemaAttributes() map[string]schema.Attribute {
	requestSchemaAttributes := map[string]schema.Attribute{
		"body": schema.StringAttribute{
			Optional: true,
		},
		"headers": schema.StringAttribute{
			Optional: true,
		},
		"method": schema.StringAttribute{
			Required: true,
		},
		"url": schema.StringAttribute{
			Required: true,
		},
	}

	return map[string]schema.Attribute{
		"register": schema.SingleNestedAttribute{
			Required: true,
			Attributes: map[string]schema.Attribute{
				"request": schema.SingleNestedAttribute{
					Required:   true,
					Attributes: requestSchemaAttributes,
				},
				"response": schema.StringAttribute{
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
			},
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
		},
		"unregister": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"request": schema.SingleNestedAttribute{
					Required:   true,
					Attributes: requestSchemaAttributes,
				},
			},
		},
	}
}
