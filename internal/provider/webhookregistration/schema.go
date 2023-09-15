package webhookregistration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *webhookRegistrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

	resp.Schema = schema.Schema{
		Description: "Webhook Resource",
		Attributes: map[string]schema.Attribute{
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
		},
	}
}
