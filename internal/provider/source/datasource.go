package source

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/schemahelpers"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sourceDataSource{}
	_ datasource.DataSourceWithConfigure = &sourceDataSource{}
)

// NewSourceDataSource returns the hookdeck_gateway_source data source.
func NewSourceDataSource() datasource.DataSource {
	return &sourceDataSource{naming: shared.Naming{Suffix: "_source"}}
}

// NewLegacySourceDataSource returns the deprecated hookdeck_source alias.
func NewLegacySourceDataSource() datasource.DataSource {
	return &sourceDataSource{naming: shared.Naming{Suffix: "_source", Legacy: true}}
}

// sourceDataSource is the datasource implementation.
type sourceDataSource struct {
	shared.ProjectScopedDataSource
	naming shared.Naming
}

// Metadata returns the datasource type name.
func (r *sourceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the data source schema.
func (r *sourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: r.naming.DataSourceDeprecation(),
		Description:        r.naming.Description("Source Data Source"),
		Attributes:         dataSourceAttributes(),
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *sourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *sourceResourceModel
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
		resp.Diagnostics.AddError("Source not found", "No source with ID "+data.ID.ValueString()+" in this project.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dataSourceAttributes() map[string]schema.Attribute {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id")
	attributes["project_id"] = shared.ProjectIDDataSourceAttribute()
	return attributes
}
