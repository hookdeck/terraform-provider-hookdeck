package project

import (
	"context"
	"fmt"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
	_ resource.ResourceWithModifyPlan  = &projectResource{}
)

// NewProjectResource returns the hookdeck_gateway_project resource.
func NewProjectResource() resource.Resource {
	return &projectResource{kind: gateway}
}

type projectResource struct {
	kind   kind
	client sdkclient.Client
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.kind.typeName(req.ProviderTypeName)
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("%s project. Managing a project requires an organization API key with `projects.write`. Deleting a project deletes everything in it.", r.kind.label),
		Attributes:  schemaAttributes(r.kind),
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := shared.ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

// organizationKey reports whether the provider has the organization API key
// projects are managed with. A project API key reads its own project only
// and answers 404 for any other, which must not be taken as "deleted".
func (r *projectResource) organizationKey(diags *diag.Diagnostics) bool {
	if r.client.Scope.KeyKind == projectscope.KeyKindOrganization {
		return true
	}
	diags.AddError("Organization API key required",
		fmt.Sprintf("%[1]s is managed with an organization API key (prefix %[2]s). The provider is configured with a project API key, which cannot create, update or delete projects. Use the %[1]s data source to read the key's own project.",
			r.kind.typeName("hookdeck"), projectscope.OrganizationKeyPrefix))
	return false
}

func (r *projectResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.organizationKey(&resp.Diagnostics)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(data.create(ctx, r.kind, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.organizationKey(&resp.Diagnostics) {
		return
	}
	var data projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, diags := data.retrieve(ctx, r.kind, r.client)
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

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(data.update(ctx, r.kind, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(data.delete(ctx, r.client)...)
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
