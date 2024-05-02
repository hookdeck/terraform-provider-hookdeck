package webhookregistration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webhookRegistrationResource{}
	_ resource.ResourceWithConfigure   = &webhookRegistrationResource{}
	_ resource.ResourceWithImportState = &webhookRegistrationResource{}
)

// NewWebhookRegistrationResource is a helper function to simplify the provider implementation.
func NewWebhookRegistrationResource() resource.Resource {
	return &webhookRegistrationResource{}
}

// webhookRegistrationResource is the resource implementation.
type webhookRegistrationResource struct {
}

// Metadata returns the resource type name.
func (r *webhookRegistrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_registration"
}

// Configure adds the provider configured client to the resource.
func (r *webhookRegistrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
}

// Create creates the resource and sets the initial Terraform state.
func (r *webhookRegistrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get data from Terraform plan
	var data *webhookRegistrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create resource
	request, err := data.ToRegisterRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error constructing custom webhook registration request", err.Error())
		return
	}
	response, err := data.DoRequest(request, false)
	if err != nil {
		resp.Diagnostics.AddError("Error registering custom webhook", err.Error())
		return
	}
	registerResponseData, err := data.ParseRegisterResponse(response)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing webhook registration response", err.Error())
		return
	}

	// Save updated data into Terraform state
	data.Refresh(registerResponseData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webhookRegistrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webhookRegistrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get data from Terraform plan
	var data *webhookRegistrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webhookRegistrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get data from Terraform state
	var data *webhookRegistrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing resource
	request, err := data.ToUnregisterRequest()
	if err != nil {
		resp.Diagnostics.AddError("Error constructing custom webhook unregistration request", err.Error())
		return
	}
	if request != nil {
		_, err = data.DoRequest(request, true)
		if err != nil {
			resp.Diagnostics.AddError("Error unregistering custom webhook", err.Error())
			return
		}
	}
}

func (r *webhookRegistrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
