package destination

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &destinationResource{}
	_ resource.ResourceWithConfigure   = &destinationResource{}
	_ resource.ResourceWithImportState = &destinationResource{}
)

// NewDestinationResource is a helper function to simplify the provider implementation.
func NewDestinationResource() resource.Resource {
	return &destinationResource{}
}

// destinationResource is the resource implementation.
type destinationResource struct {
	client sdkclient.Client
}

// Metadata returns the resource type name.
func (r *destinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

// Schema returns the resource schema.
func (r *destinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Destination Resource",
		Attributes:          schemaAttributes(),
	}
}

// Configure adds the provider configured client to the resource.
func (r *destinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(sdkclient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected sdkclient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *destinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get data from Terraform plan
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource
	destination, err := r.client.Destination.Create(context.Background(), data.ToCreatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error creating destination", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(destination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *destinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get data from Terraform state
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource value
	destination, err := r.client.Destination.Retrieve(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading destination", err.Error())
		return
	}

	// Save refreshed data into Terraform state
	data.Refresh(destination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *destinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get data from Terraform plan
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	destination, err := r.client.Destination.Update(context.Background(), data.ID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error updating destination", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(destination)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *destinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get data from Terraform state
	var data *destinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing resource
	_, err := r.client.Destination.Delete(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting destination", err.Error())
	}
}

func (r *destinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
