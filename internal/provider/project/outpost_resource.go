package project

import (
	"context"
	"maps"
	"net/http"
	"time"

	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &outpostProjectResource{}
	_ resource.ResourceWithConfigure      = &outpostProjectResource{}
	_ resource.ResourceWithImportState    = &outpostProjectResource{}
	_ resource.ResourceWithModifyPlan     = &outpostProjectResource{}
	_ resource.ResourceWithValidateConfig = &outpostProjectResource{}
)

const (
	defaultCreateTimeout = 20 * time.Minute
	defaultUpdateTimeout = 5 * time.Minute
	defaultDeleteTimeout = 10 * time.Minute
)

// NewOutpostProjectResource returns the hookdeck_outpost_project resource.
func NewOutpostProjectResource() resource.Resource {
	return &outpostProjectResource{}
}

type outpostProjectResource struct {
	client sdkclient.Client
}

func (r *outpostProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = outpost.typeName(req.ProviderTypeName)
}

func (r *outpostProjectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Outpost project. Managing a project requires an organization API key with `projects.write`. Creating a project waits until it is ready, then applies `config` and `sensitive_config`. Deleting a project deletes everything in it.",
		Attributes:  outpostSchemaAttributes(),
		Blocks:      map[string]schema.Block{"timeouts": outpostTimeoutsBlock(ctx)},
	}
}

func (r *outpostProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := shared.ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *outpostProjectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config, sensitive types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("config"), &config)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("sensitive_config"), &sensitive)...)
	if resp.Diagnostics.HasError() {
		return
	}
	elements := func(m types.Map) map[string]types.String {
		if m.IsNull() || m.IsUnknown() {
			return nil
		}
		out := map[string]types.String{}
		resp.Diagnostics.Append(m.ElementsAs(ctx, &out, false)...)
		return out
	}
	resp.Diagnostics.Append(validateConfig(elements(config), elements(sensitive))...)
}

func (r *outpostProjectResource) ModifyPlan(_ context.Context, _ resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	organizationKey(outpost, r.client, &resp.Diagnostics)
}

func (r *outpostProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data outpostProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	managed, diags := data.managed(ctx)
	resp.Diagnostics.Append(diags...)
	timeout, diags := data.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp.Diagnostics.Append(data.create(ctx, outpost, r.client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The project exists from here on. State keeps its id, so a later
	// failure taints the resource instead of leaking the project.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api := outpostAPI(r.client, data.ID.ValueString())
	if err := waitProvisioned(ctx, api, data.ID.ValueString(), timeout); err != nil {
		resp.Diagnostics.AddError("Outpost project is not ready", err.Error())
		return
	}
	if len(managed) > 0 {
		live, err := patchConfig(ctx, api, toPatchValues(managed))
		if err != nil {
			resp.Diagnostics.AddError("Error configuring Outpost project", err.Error())
			return
		}
		resp.Diagnostics.Append(data.setManaged(ctx, refreshManaged(managed, live))...)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *outpostProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !organizationKey(outpost, r.client, &resp.Diagnostics) {
		return
	}
	var data outpostProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	found, diags := data.retrieve(ctx, outpost, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	managed, diags := data.managed(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(managed) > 0 {
		live, err := getConfig(ctx, outpostAPI(r.client, data.ID.ValueString()))
		switch {
		case sdkclient.IsCode(err, http.StatusConflict, codeNotProvisioned):
			resp.Diagnostics.AddWarning("Outpost project is not ready",
				"The project is not provisioned yet, so its configuration could not be read. The configuration in state is kept as it is.")
		case err != nil:
			resp.Diagnostics.AddError("Error reading Outpost project configuration", err.Error())
			return
		default:
			resp.Diagnostics.Append(data.setManaged(ctx, refreshManaged(managed, live))...)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *outpostProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, prior outpostProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	desired, diags := data.managed(ctx)
	resp.Diagnostics.Append(diags...)
	before, diags := prior.managed(ctx)
	resp.Diagnostics.Append(diags...)
	timeout, diags := data.Timeouts.Update(ctx, defaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if data.Name.Equal(prior.Name) {
		found, diags := data.retrieve(ctx, outpost, r.client)
		resp.Diagnostics.Append(diags...)
		if !found && !resp.Diagnostics.HasError() {
			resp.Diagnostics.AddError("Error reading project", "The project no longer exists.")
		}
	} else {
		resp.Diagnostics.Append(data.update(ctx, outpost, r.client)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if !maps.Equal(desired, before) {
		api := outpostAPI(r.client, data.ID.ValueString())
		live, err := getConfig(ctx, api)
		if err != nil {
			resp.Diagnostics.AddError("Error reading Outpost project configuration", err.Error())
			return
		}
		if patch := planPatch(desired, before, live); len(patch) > 0 {
			if live, err = patchConfig(ctx, api, patch); err != nil {
				resp.Diagnostics.AddError("Error configuring Outpost project", err.Error())
				return
			}
		}
		resp.Diagnostics.Append(data.setManaged(ctx, refreshManaged(desired, live))...)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *outpostProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data outpostProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, diags := data.Timeouts.Delete(ctx, defaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	resp.Diagnostics.Append(data.delete(ctx, r.client)...)
}

func (r *outpostProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
