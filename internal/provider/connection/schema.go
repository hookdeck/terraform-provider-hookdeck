package connection

import (
	"context"

	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connection Resource",

		Attributes: map[string]schema.Attribute{
			"archived_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: "Date the connection was archived",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Date the connection was created",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Description for the connection",
			},
			"destination_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ID of a destination to bind to the connection",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: `ID of the connection`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `A unique, human-friendly name for the connection`,
			},
			"paused_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: "Date the connection was paused",
			},
			"rules": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"delay_rule": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"delay": schema.Int64Attribute{
									Required:    true,
									Description: `Delay to introduce in MS`,
								},
							},
						},
						// "filter_rule": schema.SingleNestedAttribute{
						// 	Computed: true,
						// 	PlanModifiers: []planmodifier.Object{
						// 		objectplanmodifier.RequiresReplace(),
						// 	},
						// 	Optional: true,
						// 	Attributes: map[string]schema.Attribute{
						// 		"body": schema.SingleNestedAttribute{
						// 			Computed: true,
						// 			PlanModifiers: []planmodifier.Object{
						// 				objectplanmodifier.RequiresReplace(),
						// 			},
						// 			Optional: true,
						// 			Attributes: map[string]schema.Attribute{
						// 				"str": schema.StringAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.String{
						// 						stringplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"float32": schema.NumberAttribute{
						// 					PlanModifiers: []planmodifier.Number{
						// 						numberplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"boolean": schema.BoolAttribute{
						// 					PlanModifiers: []planmodifier.Bool{
						// 						boolplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"connection_filter_property_4": schema.SingleNestedAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.Object{
						// 						objectplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional:    true,
						// 					Attributes:  map[string]schema.Attribute{},
						// 					Description: `JSON using our filter syntax to filter on request headers`,
						// 				},
						// 			},
						// 			Validators: []validator.Object{
						// 				validators.ExactlyOneChild(),
						// 			},
						// 			Description: `JSON using our filter syntax to filter on request headers`,
						// 		},
						// 		"headers": schema.SingleNestedAttribute{
						// 			Computed: true,
						// 			PlanModifiers: []planmodifier.Object{
						// 				objectplanmodifier.RequiresReplace(),
						// 			},
						// 			Optional: true,
						// 			Attributes: map[string]schema.Attribute{
						// 				"str": schema.StringAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.String{
						// 						stringplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"float32": schema.NumberAttribute{
						// 					PlanModifiers: []planmodifier.Number{
						// 						numberplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"boolean": schema.BoolAttribute{
						// 					PlanModifiers: []planmodifier.Bool{
						// 						boolplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"connection_filter_property_4": schema.SingleNestedAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.Object{
						// 						objectplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional:    true,
						// 					Attributes:  map[string]schema.Attribute{},
						// 					Description: `JSON using our filter syntax to filter on request headers`,
						// 				},
						// 			},
						// 			Validators: []validator.Object{
						// 				validators.ExactlyOneChild(),
						// 			},
						// 			Description: `JSON using our filter syntax to filter on request headers`,
						// 		},
						// 		"path": schema.SingleNestedAttribute{
						// 			Computed: true,
						// 			PlanModifiers: []planmodifier.Object{
						// 				objectplanmodifier.RequiresReplace(),
						// 			},
						// 			Optional: true,
						// 			Attributes: map[string]schema.Attribute{
						// 				"str": schema.StringAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.String{
						// 						stringplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"float32": schema.NumberAttribute{
						// 					PlanModifiers: []planmodifier.Number{
						// 						numberplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"boolean": schema.BoolAttribute{
						// 					PlanModifiers: []planmodifier.Bool{
						// 						boolplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"connection_filter_property_4": schema.SingleNestedAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.Object{
						// 						objectplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional:    true,
						// 					Attributes:  map[string]schema.Attribute{},
						// 					Description: `JSON using our filter syntax to filter on request headers`,
						// 				},
						// 			},
						// 			Validators: []validator.Object{
						// 				validators.ExactlyOneChild(),
						// 			},
						// 			Description: `JSON using our filter syntax to filter on request headers`,
						// 		},
						// 		"query": schema.SingleNestedAttribute{
						// 			Computed: true,
						// 			PlanModifiers: []planmodifier.Object{
						// 				objectplanmodifier.RequiresReplace(),
						// 			},
						// 			Optional: true,
						// 			Attributes: map[string]schema.Attribute{
						// 				"str": schema.StringAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.String{
						// 						stringplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"float32": schema.NumberAttribute{
						// 					PlanModifiers: []planmodifier.Number{
						// 						numberplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"boolean": schema.BoolAttribute{
						// 					PlanModifiers: []planmodifier.Bool{
						// 						boolplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional: true,
						// 				},
						// 				"connection_filter_property_4": schema.SingleNestedAttribute{
						// 					Computed: true,
						// 					PlanModifiers: []planmodifier.Object{
						// 						objectplanmodifier.RequiresReplace(),
						// 					},
						// 					Optional:    true,
						// 					Attributes:  map[string]schema.Attribute{},
						// 					Description: `JSON using our filter syntax to filter on request headers`,
						// 				},
						// 			},
						// 			Validators: []validator.Object{
						// 				validators.ExactlyOneChild(),
						// 			},
						// 			Description: `JSON using our filter syntax to filter on request headers`,
						// 		},
						// 	},
						// },
						"retry_rule": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"count": schema.Int64Attribute{
									Optional:    true,
									Description: `Maximum number of retries to attempt`,
								},
								"interval": schema.Int64Attribute{
									Optional:    true,
									Description: `Time in MS between each retry`,
								},
								"strategy": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf(
											"linear",
											"exponential",
										),
									},
									MarkdownDescription: `must be one of ["linear", "exponential"]` + "\n" +
										`Algorithm to use when calculating delay between retries`,
								},
							},
						},
						"transform_rule": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"transformation_id": schema.Int64Attribute{
									Required:    true,
									Description: `ID of the attached transformation object.`,
								},
							},
						},
					},
				},
			},
			"source_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `ID of a source to bind to the connection`,
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
				Description: `Date the connection was last updated`,
			},
		},
	}
}