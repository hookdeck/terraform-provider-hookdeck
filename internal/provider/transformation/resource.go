package transformation

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &transformationResource{}
	_ resource.ResourceWithConfigure   = &transformationResource{}
	_ resource.ResourceWithImportState = &transformationResource{}
	_ resource.ResourceWithModifyPlan  = &transformationResource{}
	_ resource.ResourceWithMoveState   = &transformationResource{}
)

// NewTransformationResource returns the hookdeck_gateway_transformation resource.
func NewTransformationResource() resource.Resource {
	return &transformationResource{naming: shared.Naming{Suffix: "_transformation"}}
}

// NewLegacyTransformationResource returns the deprecated hookdeck_transformation alias.
func NewLegacyTransformationResource() resource.Resource {
	return &transformationResource{naming: shared.Naming{Suffix: "_transformation", Legacy: true}}
}

// transformationResource is the resource implementation.
type transformationResource struct {
	shared.ProjectScopedResource
	naming shared.Naming
}

// Metadata returns the resource type name.
func (r *transformationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *transformationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *transformationResource) schema() schema.Schema {
	return schema.Schema{
		DeprecationMessage: r.naming.DeprecationMessage(),
		Description:        "Transformation Resource",
		Attributes:         schemaAttributes(),
	}
}

// MoveState accepts state from the v2 name via a moved block.
func (r *transformationResource) MoveState(_ context.Context) []resource.StateMover {
	if r.naming.Legacy {
		return nil
	}
	return []resource.StateMover{shared.RenamedStateMover(r.naming.LegacyTypeName(), r.schema())}
}

// Create creates the resource and sets the initial Terraform state.
func (r *transformationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForCreate(data.ProjectID, &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Create(ctx, client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *transformationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(data.ProjectID, data.TeamID), &resp.Diagnostics)
	if !ok {
		return
	}
	found, diags := data.Retrieve(ctx, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *transformationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state *transformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(state.ProjectID, state.TeamID), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Update(ctx, client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *transformationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *transformationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(data.ProjectID, data.TeamID), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Delete(ctx, client)...)
}

func (r *transformationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.ImportProjectState(ctx, "id", req, resp)
}
