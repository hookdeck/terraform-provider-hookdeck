package connection

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/schemahelpers"

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
	shared.ProjectScopedDataSource
	naming shared.Naming
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

// Read refreshes the Terraform state with the latest data.
func (r *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.Client(data.ProjectID, &resp.Diagnostics)
	if !ok {
		return
	}
	found, diags := data.Retrieve(ctx, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Connection not found", "No connection with ID "+data.ID.ValueString()+" in this project.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dataSourceAttributes() map[string]schema.Attribute {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id")
	attributes["project_id"] = shared.ProjectIDDataSourceAttribute()
	return attributes
}
