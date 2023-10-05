package transformation

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &transformationResource{}
	_ resource.ResourceWithConfigure   = &transformationResource{}
	_ resource.ResourceWithImportState = &transformationResource{}
)

// NewTransformationResource is a helper function to simplify the provider implementation.
func NewTransformationResource() resource.Resource {
	return &transformationResource{}
}

// transformationResource is the resource implementation.
type transformationResource struct {
	client hookdeckClient.Client
}

// Metadata returns the resource type name.
func (r *transformationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transformation"
}

// Configure adds the provider configured client to the resource.
func (r *transformationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *transformationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get data from Terraform plan
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource
	transformation, err := r.client.Transformation.Create(context.Background(), data.ToCreatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error creating transformation", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(transformation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *transformationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get data from Terraform state
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource value
	transformation, err := r.client.Transformation.Retrieve(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading transformation", err.Error())
		return
	}

	// Save refreshed data into Terraform state
	data.Refresh(transformation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *transformationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get data from Terraform plan
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	transformation, err := r.client.Transformation.Update(context.Background(), data.ID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error updating transformation", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(transformation)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *transformationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get data from Terraform state
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing resource
	// TODO: use delete once the endpoint is ready
	// _, err := r.client.Transformation.Delete(context.Background(), data.ID.ValueString())
	// if err != nil {
	// 	resp.Diagnostics.AddError("Error deleting source", err.Error())
	// }

	// TODO: remove later
	// for now, we'll update the transformation to a random ID in this template `deleted-${transformation_name}-${random}`
	// so users can still create a new transformation with the old name
	length := 10 // length of random key
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length+2)
	rand.Read(b)
	randomizedName := "deleted-" + data.Name.ValueString() + "-" + fmt.Sprintf("%x", b)[2:length+2]
	_, err := r.client.Transformation.Update(context.Background(), data.ID.ValueString(), &hookdeck.TransformationUpdateRequest{
		Name: hookdeck.OptionalOrNull(&randomizedName),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source", err.Error())
	}
}

func (r *transformationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
