package destination

import (
	"context"

	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
			// "auth_method": schema.SingleNestedAttribute{
			// 	Optional: true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"api_key": schema.SingleNestedAttribute{
			// 			Computed: true,
			// 			Optional: true,
			// 			Attributes: map[string]schema.Attribute{
			// 				"config": schema.SingleNestedAttribute{
			// 					Computed: true,
			// 					Optional: true,
			// 					Attributes: map[string]schema.Attribute{
			// 						"api_key": schema.StringAttribute{
			// 							Required:    true,
			// 							Description: `API key for the API key auth`,
			// 						},
			// 						"key": schema.StringAttribute{
			// 							Required:    true,
			// 							Description: `Key for the API key auth`,
			// 						},
			// 						"to": schema.StringAttribute{
			// 							Computed:   true,
			// 							Optional:   true,
			// 							Validators: []validator.String{stringvalidator.OneOf("header", "query")},
			// 							MarkdownDescription: `must be one of ["header", "query"]` + "\n" +
			// 								`Whether the API key should be sent as a header or a query parameter`,
			// 						},
			// 					},
			// 					Description: `API key config for the destination's auth method`,
			// 				},
			// 				"type": schema.StringAttribute{
			// 					Required:   true,
			// 					Validators: []validator.String{stringvalidator.OneOf("API_KEY")},
			// 					MarkdownDescription: `must be one of ["API_KEY"]` + "\n" +
			// 						`Type of auth method`,
			// 				},
			// 			},
			// 			Description: `API Key`,
			// 		},
			// 		"basic_auth": schema.SingleNestedAttribute{
			// 			Computed: true,
			// 			PlanModifiers: []planmodifier.Object{
			// 				objectplanmodifier.RequiresReplace(),
			// 			},
			// 			Optional: true,
			// 			Attributes: map[string]schema.Attribute{
			// 				"config": schema.SingleNestedAttribute{
			// 					Computed: true,
			// 					PlanModifiers: []planmodifier.Object{
			// 						objectplanmodifier.RequiresReplace(),
			// 					},
			// 					Optional: true,
			// 					Attributes: map[string]schema.Attribute{
			// 						"password": schema.StringAttribute{
			// 							PlanModifiers: []planmodifier.String{
			// 								stringplanmodifier.RequiresReplace(),
			// 							},
			// 							Required:    true,
			// 							Description: `Password for basic auth`,
			// 						},
			// 						"username": schema.StringAttribute{
			// 							PlanModifiers: []planmodifier.String{
			// 								stringplanmodifier.RequiresReplace(),
			// 							},
			// 							Required:    true,
			// 							Description: `Username for basic auth`,
			// 						},
			// 					},
			// 					Description: `Basic auth config for the destination's auth method`,
			// 				},
			// 				"type": schema.StringAttribute{
			// 					PlanModifiers: []planmodifier.String{
			// 						stringplanmodifier.RequiresReplace(),
			// 					},
			// 					Required: true,
			// 					Validators: []validator.String{
			// 						stringvalidator.OneOf(
			// 							"BASIC_AUTH",
			// 						),
			// 					},
			// 					MarkdownDescription: `must be one of ["BASIC_AUTH"]` + "\n" +
			// 						`Type of auth method`,
			// 				},
			// 			},
			// 			Description: `Basic Auth`,
			// 		},
			// 		"bearer_token": schema.SingleNestedAttribute{
			// 			Computed: true,
			// 			PlanModifiers: []planmodifier.Object{
			// 				objectplanmodifier.RequiresReplace(),
			// 			},
			// 			Optional: true,
			// 			Attributes: map[string]schema.Attribute{
			// 				"config": schema.SingleNestedAttribute{
			// 					Computed: true,
			// 					PlanModifiers: []planmodifier.Object{
			// 						objectplanmodifier.RequiresReplace(),
			// 					},
			// 					Optional: true,
			// 					Attributes: map[string]schema.Attribute{
			// 						"token": schema.StringAttribute{
			// 							PlanModifiers: []planmodifier.String{
			// 								stringplanmodifier.RequiresReplace(),
			// 							},
			// 							Required:    true,
			// 							Description: `Token for the bearer token auth`,
			// 						},
			// 					},
			// 					Description: `Bearer token config for the destination's auth method`,
			// 				},
			// 				"type": schema.StringAttribute{
			// 					PlanModifiers: []planmodifier.String{
			// 						stringplanmodifier.RequiresReplace(),
			// 					},
			// 					Required: true,
			// 					Validators: []validator.String{
			// 						stringvalidator.OneOf(
			// 							"BEARER_TOKEN",
			// 						),
			// 					},
			// 					MarkdownDescription: `must be one of ["BEARER_TOKEN"]` + "\n" +
			// 						`Type of auth method`,
			// 				},
			// 			},
			// 			Description: `Bearer Token`,
			// 		},
			// 		"custom_signature": schema.SingleNestedAttribute{
			// 			Computed: true,
			// 			PlanModifiers: []planmodifier.Object{
			// 				objectplanmodifier.RequiresReplace(),
			// 			},
			// 			Optional: true,
			// 			Attributes: map[string]schema.Attribute{
			// 				"config": schema.SingleNestedAttribute{
			// 					PlanModifiers: []planmodifier.Object{
			// 						objectplanmodifier.RequiresReplace(),
			// 					},
			// 					Required: true,
			// 					Attributes: map[string]schema.Attribute{
			// 						"key": schema.StringAttribute{
			// 							PlanModifiers: []planmodifier.String{
			// 								stringplanmodifier.RequiresReplace(),
			// 							},
			// 							Required:    true,
			// 							Description: `Key for the custom signature auth`,
			// 						},
			// 						"signing_secret": schema.StringAttribute{
			// 							Computed: true,
			// 							PlanModifiers: []planmodifier.String{
			// 								stringplanmodifier.RequiresReplace(),
			// 							},
			// 							Optional:    true,
			// 							Description: `Signing secret for the custom signature auth. If left empty a secret will be generated for you.`,
			// 						},
			// 					},
			// 					Description: `Custom signature config for the destination's auth method`,
			// 				},
			// 				"type": schema.StringAttribute{
			// 					PlanModifiers: []planmodifier.String{
			// 						stringplanmodifier.RequiresReplace(),
			// 					},
			// 					Required: true,
			// 					Validators: []validator.String{
			// 						stringvalidator.OneOf(
			// 							"CUSTOM_SIGNATURE",
			// 						),
			// 					},
			// 					MarkdownDescription: `must be one of ["CUSTOM_SIGNATURE"]` + "\n" +
			// 						`Type of auth method`,
			// 				},
			// 			},
			// 			Description: `Custom Signature`,
			// 		},
			// 		"hookdeck_signature": schema.SingleNestedAttribute{
			// 			Computed: true,
			// 			PlanModifiers: []planmodifier.Object{
			// 				objectplanmodifier.RequiresReplace(),
			// 			},
			// 			Optional: true,
			// 			Attributes: map[string]schema.Attribute{
			// 				"config": schema.SingleNestedAttribute{
			// 					Computed: true,
			// 					PlanModifiers: []planmodifier.Object{
			// 						objectplanmodifier.RequiresReplace(),
			// 					},
			// 					Optional:    true,
			// 					Attributes:  map[string]schema.Attribute{},
			// 					Description: `Empty config for the destination's auth method`,
			// 				},
			// 				"type": schema.StringAttribute{
			// 					PlanModifiers: []planmodifier.String{
			// 						stringplanmodifier.RequiresReplace(),
			// 					},
			// 					Required: true,
			// 					Validators: []validator.String{
			// 						stringvalidator.OneOf(
			// 							"HOOKDECK_SIGNATURE",
			// 						),
			// 					},
			// 					MarkdownDescription: `must be one of ["HOOKDECK_SIGNATURE"]` + "\n" +
			// 						`Type of auth method`,
			// 				},
			// 			},
			// 			Description: `Hookdeck Signature`,
			// 		},
			// 	},
			// 	Validators: []validator.Object{
			// 		validators.ExactlyOneChild(),
			// 	},
			// 	Description: `Config for the destination's auth method`,
			// },
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
			"rate_limit": schema.Int64Attribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: `Limit event attempts to receive per period. Max value is workspace plan's max attempts thoughput.`,
			},
			"rate_limit_period": schema.StringAttribute{
				Computed: true,
				Optional: true,
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
