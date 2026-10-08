package destination

import (
	"context"
	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &destinationResource{}
	_ resource.ResourceWithConfigure      = &destinationResource{}
	_ resource.ResourceWithImportState    = &destinationResource{}
	_ resource.ResourceWithModifyPlan     = &destinationResource{}
	_ resource.ResourceWithMoveState      = &destinationResource{}
	_ resource.ResourceWithValidateConfig = &destinationResource{}
)

// NewDestinationResource returns the hookdeck_gateway_destination resource.
func NewDestinationResource() resource.Resource {
	return newDestinationResource(false)
}

// NewLegacyDestinationResource returns the deprecated hookdeck_destination alias.
func NewLegacyDestinationResource() resource.Resource {
	return newDestinationResource(true)
}

func newDestinationResource(legacy bool) *destinationResource {
	return &destinationResource{
		ProjectScopedResource: shared.ProjectScopedResource{ListPath: "/destinations"},
		naming:                shared.Naming{Suffix: "_destination", Legacy: legacy},
	}
}

// destinationResource is the resource implementation.
type destinationResource struct {
	shared.ProjectScopedResource
	naming shared.Naming
}

// Metadata returns the resource type name.
func (r *destinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.naming.TypeName(req.ProviderTypeName)
}

// Schema returns the resource schema.
func (r *destinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = r.schema()
}

func (r *destinationResource) schema() schema.Schema {
	return resourceSchema(r.naming)
}

func resourceSchema(naming shared.Naming) schema.Schema {
	return schema.Schema{
		DeprecationMessage:  naming.ResourceDeprecation(),
		MarkdownDescription: naming.Description("Destination Resource"),
		Attributes:          naming.WithTeamID(schemaAttributes()),
	}
}

// MoveState accepts state from the v2 name via a moved block.
func (r *destinationResource) MoveState(_ context.Context) []resource.StateMover {
	if r.naming.Legacy {
		return nil
	}
	return []resource.StateMover{shared.RenamedStateMover(r.naming.LegacyTypeName(), resourceSchema(r.naming.AsLegacy()), r.schema())}
}

// Create creates the resource and sets the initial Terraform state.
func (r *destinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
func (r *destinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
func (r *destinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	resp.Diagnostics.Append(data.Update(ctx, client, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setModel(ctx, &resp.State, r.naming.Legacy, data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *destinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

func (r *destinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.ImportProjectState(ctx, "id", req, resp)
}

// ValidateConfig runs plan-time checks on the resource configuration.
func (r *destinationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data destinationResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateConfigForAPIVersion20260901(data.Config)...)
}
