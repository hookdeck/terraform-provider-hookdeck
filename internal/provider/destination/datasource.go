package destination

import (
	"context"
	"fmt"
	schemaHelpers "terraform-provider-hookdeck/internal/schemahelpers"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &destinationDataSource{}
	_ datasource.DataSourceWithConfigure = &destinationDataSource{}
)

// NewDestinationDataSource is a helper function to simplify the provider implementation.
func NewDestinationDataSource() datasource.DataSource {
	return &destinationDataSource{}
}

// destinationDataSource is the datasource implementation.
type destinationDataSource struct {
	client hookdeckClient.Client
}

// Metadata returns the datasource type name.
func (r *destinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

// Schema returns the data source schema.
func (r *destinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Destination Data Source",
		Attributes:  schemaHelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id"),
	}
}

// Configure adds the provider configured client to the datasource.
func (r *destinationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (r *destinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get data from Terraform state
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed datasource value
	destination, err := r.client.Destination.Retrieve(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading destination", err.Error())
		return
	}

	// Save refreshed data into Terraform state
	data.Refresh(destination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
