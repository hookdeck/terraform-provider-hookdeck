package connection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &connectionResource{}
	_ resource.ResourceWithConfigure   = &connectionResource{}
	_ resource.ResourceWithImportState = &connectionResource{}
)

// NewConnectionResource is a helper function to simplify the provider implementation.
func NewConnectionResource() resource.Resource {
	return &connectionResource{}
}

// connectionResource is the resource implementation.
type connectionResource struct {
	client hookdeckClient.Client
}

// Metadata returns the resource type name.
func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

// Schema returns the resource schema.
func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connection Resource",
		Attributes:          schemaAttributes(),
	}
}

// Configure adds the provider configured client to the resource.
func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hookdeckClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Connection Configure Type",
			fmt.Sprintf("Expected *hookdeckClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = *client
}

// Create creates the resource and sets the initial Terraform state.
func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get data from Terraform plan
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource
	connection, err := r.client.Connection.Create(context.Background(), data.ToCreatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error creating connection", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(connection)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get data from Terraform state
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource value
	connection, err := r.client.Connection.Retrieve(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading connection", err.Error())
		return
	}

	// Save refreshed data into Terraform state
	data.Refresh(connection)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get data from Terraform plan
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	connection, err := r.client.Connection.Update(context.Background(), data.ID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error updating connection", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(connection)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get data from Terraform state
	var data *connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing resource
	_, err := r.client.Connection.Delete(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting connection", err.Error())
	}
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
