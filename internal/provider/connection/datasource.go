package connection

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/schemahelpers"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &connectionDataSource{}
	_ datasource.DataSourceWithConfigure = &connectionDataSource{}
)

// NewConnectionDataSource is a helper function to simplify the provider implementation.
func NewConnectionDataSource() datasource.DataSource {
	return &connectionDataSource{}
}

// connectionDataSource is the datasource implementation.
type connectionDataSource struct {
	client sdkclient.Client
}

// Metadata returns the datasource type name.
func (r *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

// Schema returns the data source schema.
func (r *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Connection Data Source",
		Attributes:  schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id"),
	}
}

// Configure adds the provider configured client to the datasource.
func (r *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(sdkclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected sdkclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Read refreshes the Terraform state with the latest data.
func (r *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get data from Terraform state
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed datasource value
	resp.Diagnostics.Append(data.Retrieve(ctx, &r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save refreshed data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
