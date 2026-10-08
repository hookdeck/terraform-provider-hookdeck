package connection

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &connectionResource{}
	_ resource.ResourceWithConfigure   = &connectionResource{}
	_ resource.ResourceWithImportState = &connectionResource{}
	_ resource.ResourceWithModifyPlan  = &connectionResource{}
	_ resource.ResourceWithMoveState   = &connectionResource{}
)

// NewConnectionResource returns the hookdeck_gateway_connection resource.
func NewConnectionResource() resource.Resource {
	return newConnectionResource(false)
}

// NewLegacyConnectionResource returns the deprecated hookdeck_connection alias.
func NewLegacyConnectionResource() resource.Resource {
	return newConnectionResource(true)
}

func newConnectionResource(legacy bool) *connectionResource {
	return &connectionResource{
		ProjectScopedResource: shared.ProjectScopedResource{ListPath: "/connections"},
		naming:                shared.Naming{Suffix: "_connection", Legacy: legacy},
	}
}

// connectionResource is the resource implementation.
type connectionResource struct {
	shared.ProjectScopedResource
	naming shared.Naming
}

// Metadata returns the resource type name.
func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *connectionResource) schema() schema.Schema {
	return resourceSchema(r.naming)
}

func resourceSchema(naming shared.Naming) schema.Schema {
	return schema.Schema{
		DeprecationMessage:  naming.ResourceDeprecation(),
		Version:             1,
		MarkdownDescription: naming.Description("Connection Resource"),
		Attributes:          naming.WithTeamID(schemaAttributes()),
	}
}

// MoveState accepts state from the v2 name via a moved block.
func (r *connectionResource) MoveState(_ context.Context) []resource.StateMover {
	if r.naming.Legacy {
		return nil
	}
	return []resource.StateMover{shared.RenamedStateMover(r.naming.LegacyTypeName(), resourceSchema(r.naming.AsLegacy()), r.schema())}
}

// Create creates the resource and sets the initial Terraform state.
func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	data, _, diags := getModel(ctx, req.Plan, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
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

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data, teamID, diags := getModel(ctx, req.State, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(data.ProjectID, teamID), &resp.Diagnostics)
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

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	data, _, diags := getModel(ctx, req.Plan, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, teamID, diags := getModel(ctx, req.State, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(state.ProjectID, teamID), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Update(ctx, client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	data, teamID, diags := getModel(ctx, req.State, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(data.ProjectID, teamID), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(data.Delete(ctx, client)...)
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.ImportProjectState(ctx, "id", req, resp)
}
