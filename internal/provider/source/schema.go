package source

import (
	"context"

	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *sourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultAllowedHttpMethods, _ := types.ListValueFrom(context.Background(), types.StringType, []string{"POST", "PUT", "PATCH", "DELETE"})

	resp.Schema = schema.Schema{
		Description: "Source Resource",
		Attributes: map[string]schema.Attribute{
			"allowed_http_methods": schema.ListAttribute{
				Computed:    true,
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(
						"GET",
						"POST",
						"PUT",
						"PATCH",
						"DELETE",
					)),
				},
				Default:     listdefault.StaticValue(defaultAllowedHttpMethods),
				Description: "List of allowed HTTP methods. Defaults to PUT, POST, PATCH, DELETE.",
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
			"custom_response": schema.SingleNestedAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"body": schema.StringAttribute{
						Required:    true,
						Description: "Body of the custom response",
					},
					"content_type": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"json",
								"text",
								"xml",
							),
						},
						MarkdownDescription: "must be one of [json, text, xml]" + "\n" +
							"Content type of the custom response",
					},
				},
				Description: "Custom response object",
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
				Optional: true,
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
		},
	}
}
