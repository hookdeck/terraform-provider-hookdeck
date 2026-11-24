package source

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                 = &sourceResource{}
	_ resource.ResourceWithConfigure    = &sourceResource{}
	_ resource.ResourceWithImportState  = &sourceResource{}
	_ resource.ResourceWithModifyPlan   = &sourceResource{}
	_ resource.ResourceWithMoveState    = &sourceResource{}
	_ resource.ResourceWithUpgradeState = &sourceResource{}
)

// NewSourceResource returns the hookdeck_gateway_source resource.
func NewSourceResource() resource.Resource {
	return newSourceResource(false)
}

// NewLegacySourceResource returns the deprecated hookdeck_source alias.
func NewLegacySourceResource() resource.Resource {
	return newSourceResource(true)
}

func newSourceResource(legacy bool) *sourceResource {
	return &sourceResource{
		ProjectScopedResource: shared.ProjectScopedResource{ListPath: "/sources"},
		naming:                shared.Naming{Suffix: "_source", Legacy: legacy},
	}
}

// sourceResource is the resource implementation.
type sourceResource struct {
	shared.ProjectScopedResource
	naming shared.Naming
}

// Metadata returns the resource type name.
func (r *sourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *sourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *sourceResource) schema() schema.Schema {
	return resourceSchema(r.naming)
}

func resourceSchema(naming shared.Naming) schema.Schema {
	return schema.Schema{
		Version:            shared.ConfigSchemaVersion,
		DeprecationMessage: naming.ResourceDeprecation(),
		Description:        naming.Description("Source Resource"),
		Attributes:         naming.WithTeamID(schemaAttributes()),
	}
}

// UpgradeState accepts state written before config was read back from
// Hookdeck. Nothing in it changes.
func (r *sourceResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return shared.UnchangedStateUpgraders(r.schema())
}

// MoveState accepts state from the v2 name via a moved block.
func (r *sourceResource) MoveState(_ context.Context) []resource.StateMover {
	if r.naming.Legacy {
		return nil
	}
	return []resource.StateMover{shared.RenamedStateMover(r.naming.LegacyTypeName(), resourceSchema(r.naming.AsLegacy()), r.schema())}
}

// ModifyPlan checks the project, then turns a plan that changes only computed
// values back into the state, so a config that already matches Hookdeck plans
// no update.
func (r *sourceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.ProjectScopedResource.ModifyPlan(ctx, req, resp)
	if resp.Diagnostics.HasError() || len(resp.RequiresReplace) > 0 {
		return
	}
	resp.Diagnostics.Append(shared.KeepUnchangedPlan(ctx, req.State, &resp.Plan)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *sourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	data, _, diags := getModel(ctx, req.Plan, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForCreate(data.ProjectID, &resp.Diagnostics)
	if !ok {
		return
	}
	var userConfig jsontypes.Normalized
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("config"), &userConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(data.Create(ctx, client, userConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(shared.SetManagedConfigPaths(ctx, resp.Private, userConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *sourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data, teamID, diags := getModel(ctx, req.State, r.naming.Legacy)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, ok := r.ClientForState(shared.StoredProject(data.ProjectID, teamID), &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(shared.SeedManagedConfigPaths(ctx, resp.Private, data.Config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, diags := data.Retrieve(ctx, client, true)
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
func (r *sourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	var userConfig jsontypes.Normalized
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("config"), &userConfig)...)
	managed, _, diags := shared.ManagedConfigPaths(ctx, req.Private, state.Config.StringValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A plan that keeps the stored config (nothing to change in it, or
	// ignore_changes) sends no config and leaves the recorded keys as they are.
	keepConfig := shared.ConfigUnchanged(data.Config, state.Config, data.Type, state.Type)
	if keepConfig {
		userConfig = jsontypes.NewNormalizedNull()
		managed = nil
	}
	resp.Diagnostics.Append(data.Update(ctx, client, userConfig, managed, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !keepConfig {
		resp.Diagnostics.Append(shared.SetManagedConfigPaths(ctx, resp.Private, userConfig)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *sourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *sourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.ImportProjectState(ctx, "id", req, resp)
	shared.MarkConfigUnrecorded(ctx, resp)
}
