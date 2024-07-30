package source

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sourceDataSource{}
	_ datasource.DataSourceWithConfigure = &sourceDataSource{}
)

// NewSourceDataSource is a helper function to simplify the provider implementation.
func NewSourceDataSource() datasource.DataSource {
	return &sourceDataSource{}
}

// sourceDataSource is the datasource implementation.
type sourceDataSource struct {
	client hookdeckClient.Client
}

// Metadata returns the datasource type name.
func (r *sourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

// Configure adds the provider configured client to the datasource.
func (r *sourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hookdeckClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hookdeckClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = *client
}

// Read refreshes the Terraform state with the latest data.
func (r *sourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get data from Terraform state
	var data *sourceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed datasource value
	source, err := r.client.Source.Retrieve(context.Background(), data.ID.ValueString(), &hookdeck.SourceRetrieveRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Error reading source", err.Error())
		return
	}

	// Save refreshed data into Terraform state
	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Description: "List of allowed HTTP methods. Defaults to PUT, POST, PATCH, DELETE.",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: "Date the source was created",
			},
			"custom_response": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
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
				Computed:    true,
				Optional:    true,
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
				Required:    true,
				Description: "ID of the source",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "A unique, human-friendly name for the source",
			},
			"team_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the workspace",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date the source was last updated",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "A unique URL that must be supplied to your webhook's provider",
			},
		},
	}
}
