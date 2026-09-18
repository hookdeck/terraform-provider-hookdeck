package connection

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/provider/shared"
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

// NewConnectionDataSource returns the hookdeck_gateway_connection data source.
func NewConnectionDataSource() datasource.DataSource {
	return &connectionDataSource{naming: shared.Naming{Suffix: "_connection"}}
}

// NewLegacyConnectionDataSource returns the deprecated hookdeck_connection alias.
func NewLegacyConnectionDataSource() datasource.DataSource {
	return &connectionDataSource{naming: shared.Naming{Suffix: "_connection", Legacy: true}}
}

// connectionDataSource is the datasource implementation.
type connectionDataSource struct {
	naming shared.Naming
	client sdkclient.Client
}

// Metadata returns the datasource type name.
func (r *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the data source schema.
func (r *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: r.naming.DeprecationMessage(),
		Description:        "Connection Data Source",
		Attributes:         dataSourceAttributes(),
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
	client, clientDiags := shared.ClientFor(ctx, r.client, data.ProjectID)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := data.Retrieve(ctx, &client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save refreshed data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dataSourceAttributes() map[string]schema.Attribute {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id")
	attributes["project_id"] = shared.ProjectIDDataSourceAttribute()
	return attributes
}
