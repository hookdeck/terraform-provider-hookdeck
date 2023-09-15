package destination

import (
	"context"

	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *destinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Destination Resource",

		Attributes: map[string]schema.Attribute{
			"archived_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: `Date the destination was archived`,
			},
			"auth_method": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"api_key": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"api_key": schema.StringAttribute{
								Required:    true,
								Sensitive:   true,
								Description: `API key for the API key auth`,
							},
							"key": schema.StringAttribute{
								Required:    true,
								Description: `Key for the API key auth`,
							},
							"to": schema.StringAttribute{
								Computed:   true,
								Optional:   true,
								Validators: []validator.String{stringvalidator.OneOf("header", "query")},
								Default:    stringdefault.StaticString("header"),
								MarkdownDescription: `must be one of ["header", "query"]` + "\n" +
									`Whether the API key should be sent as a header or a query parameter`,
							},
						},
						Description: `API Key`,
					},
					"basic_auth": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"password": schema.StringAttribute{
								Required:    true,
								Sensitive:   true,
								Description: `Password for basic auth`,
							},
							"username": schema.StringAttribute{
								Required:    true,
								Description: `Username for basic auth`,
							},
						},
						Description: `Basic Auth`,
					},
					"bearer_token": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"token": schema.StringAttribute{
								Required:    true,
								Sensitive:   true,
								Description: `Token for the bearer token auth`,
							},
						},
						Description: `Bearer Token`,
					},
					"custom_signature": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"key": schema.StringAttribute{
								Required:    true,
								Description: `Key for the custom signature auth`,
							},
							"signing_secret": schema.StringAttribute{
								Computed:    true,
								Optional:    true,
								Sensitive:   true,
								Description: `Signing secret for the custom signature auth. If left empty a secret will be generated for you.`,
							},
						},
						Description: `Custom Signature`,
					},
					"hookdeck_signature": schema.SingleNestedAttribute{
						Optional:    true,
						Attributes:  map[string]schema.Attribute{},
						Description: `Hookdeck Signature`,
					},
				},
				Validators: []validator.Object{
					validators.ExactlyOneChild(),
				},
				Default: objectdefault.StaticValue(
					types.ObjectValueMust(
						destinationAuthMethodConfig{}.attrTypes(),
						destinationAuthMethodConfig{}.defaultValue(),
					),
				),
				Description: `Config for the destination's auth method`,
			},
			"cli_path": schema.StringAttribute{
				Optional:    true,
				Description: `Path for the CLI destination`,
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
			"http_method": schema.StringAttribute{
				// Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"GET",
						"POST",
						"PUT",
						"PATCH",
						"DELETE",
					),
				},
				MarkdownDescription: `must be one of ["GET", "POST", "PUT", "PATCH", "DELETE"]` + "\n" +
					`HTTP method used on requests sent to the destination, overrides the method used on requests sent to the source.`,
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
			"path_forwarding_disabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"rate_limit": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"limit": schema.Int64Attribute{
						Required: true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Description: `Limit event attempts to receive per period. Max value is workspace plan's max attempts thoughput.`,
					},
					"period": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"second",
								"minute",
								"hour",
							),
						},
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: `must be one of ["second", "minute", "hour"]` + "\n" +
							`Period to rate limit attempts`,
					},
				},
				Description: "Rate limit",
			},
			"team_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `ID of the workspace`,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: `Date the destination was last updated`,
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Description: `HTTP endpoint of the destination`,
			},
		},
	}
}
