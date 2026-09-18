package sourceauth

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &sourceAuthResource{}
	_ resource.ResourceWithConfigure   = &sourceAuthResource{}
	_ resource.ResourceWithImportState = &sourceAuthResource{}
	_ resource.ResourceWithMoveState   = &sourceAuthResource{}
)

// NewSourceAuthResource returns the hookdeck_gateway_source_auth resource.
func NewSourceAuthResource() resource.Resource {
	return &sourceAuthResource{naming: shared.Naming{Suffix: "_source_auth"}}
}

// NewLegacySourceAuthResource returns the deprecated hookdeck_source_auth alias.
func NewLegacySourceAuthResource() resource.Resource {
	return &sourceAuthResource{naming: shared.Naming{Suffix: "_source_auth", Legacy: true}}
}

// sourceAuthResource is the resource implementation.
type sourceAuthResource struct {
	naming shared.Naming
	client sdkclient.Client
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

// Configure adds the provider configured client to the resource.
func (r *sourceAuthResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *sourceAuthResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *sourceAuthResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, clientDiags := shared.ClientFor(ctx, r.client, data.ProjectID)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := data.Create(ctx, &client)
	resp.Diagnostics.Append(diags...)
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

	client, clientDiags := shared.ClientFor(ctx, r.client, data.ProjectID)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := data.Retrieve(ctx, &client)
	if shared.IsNotFoundDiagnostics(diags) {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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

	client, clientDiags := shared.ClientFor(ctx, r.client, data.ProjectID)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := data.Update(ctx, &client)
	resp.Diagnostics.Append(diags...)
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

	client, clientDiags := shared.ClientFor(ctx, r.client, data.ProjectID)
	resp.Diagnostics.Append(clientDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := data.Delete(ctx, &client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceAuthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	shared.ImportState(ctx, "source_id", req, resp)
}
