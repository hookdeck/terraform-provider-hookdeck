package destination

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
	_ datasource.DataSource              = &destinationDataSource{}
	_ datasource.DataSourceWithConfigure = &destinationDataSource{}
)

// NewDestinationDataSource returns the hookdeck_gateway_destination data source.
func NewDestinationDataSource() datasource.DataSource {
	return &destinationDataSource{naming: shared.Naming{Suffix: "_destination"}}
}

// NewLegacyDestinationDataSource returns the deprecated hookdeck_destination alias.
func NewLegacyDestinationDataSource() datasource.DataSource {
	return &destinationDataSource{naming: shared.Naming{Suffix: "_destination", Legacy: true}}
}

// destinationDataSource is the datasource implementation.
type destinationDataSource struct {
	naming shared.Naming
	client sdkclient.Client
}

// Metadata returns the datasource type name.
func (r *destinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the data source schema.
func (r *destinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		DeprecationMessage: r.naming.DeprecationMessage(),
		Description:        "Destination Data Source",
		Attributes:         dataSourceAttributes(),
	}
}

// Configure adds the provider configured client to the datasource.
func (r *destinationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (r *destinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dataSourceAttributes() map[string]schema.Attribute {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id")
	attributes["project_id"] = shared.ProjectIDDataSourceAttribute()
	return attributes
}
