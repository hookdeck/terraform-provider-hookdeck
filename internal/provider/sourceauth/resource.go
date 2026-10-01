package sourceauth

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &sourceAuthResource{}
	_ resource.ResourceWithConfigure   = &sourceAuthResource{}
	_ resource.ResourceWithImportState = &sourceAuthResource{}
	_ resource.ResourceWithModifyPlan  = &sourceAuthResource{}
	_ resource.ResourceWithMoveState   = &sourceAuthResource{}
)

func newSourceAuthResource(legacy bool) *sourceAuthResource {
	return &sourceAuthResource{
		ProjectScopedResource: shared.ProjectScopedResource{ParentAttribute: "source_id"},
		naming:                shared.Naming{Suffix: "_source_auth", Legacy: legacy},
	}
}

// NewSourceAuthResource returns the hookdeck_gateway_source_auth resource.
func NewSourceAuthResource() resource.Resource {
	return newSourceAuthResource(false)
}

// NewLegacySourceAuthResource returns the deprecated hookdeck_source_auth alias.
func NewLegacySourceAuthResource() resource.Resource {
	return newSourceAuthResource(true)
}

// sourceAuthResource is the resource implementation.
type sourceAuthResource struct {
	shared.ProjectScopedResource
	naming shared.Naming
}

// Metadata returns the resource type name.
func (r *sourceAuthResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *sourceAuthResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *sourceAuthResource) schema() schema.Schema {
	return schema.Schema{
		DeprecationMessage: r.naming.DeprecationMessage(),
		Description:        "Source Auth Resource",
		Attributes:         schemaAttributes(),
	}
}

// MoveState accepts state from the v2 name via a moved block.
func (r *sourceAuthResource) MoveState(_ context.Context) []resource.StateMover {
	if r.naming.Legacy {
		return nil
	}
	return []resource.StateMover{shared.RenamedStateMover(r.naming.LegacyTypeName(), r.schema())}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sourceAuthResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *sourceAuthResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForCreate(data.ProjectID, &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Update(ctx, client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *sourceAuthResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *sourceAuthResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// State written by v2 records no project. Without a provider project
	// there is nothing to read with until the first apply stores project_id.
	if data.ProjectID.ValueString() == "" && !r.SingleProject() {
		return
	}

	client, ok := r.ClientForState(data.ProjectID.ValueString(), &resp.Diagnostics)
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
func (r *sourceAuthResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *sourceAuthResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The plan's project, not the state's: the auth follows its source,
	// and a new source_id can be in another project.
	client, ok := r.ClientForCreate(data.ProjectID, &resp.Diagnostics)
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
func (r *sourceAuthResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *sourceAuthResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(data.ProjectID.ValueString(), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Delete(ctx, client)...)
}

func (r *sourceAuthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.ImportProjectState(ctx, "source_id", req, resp)
}
