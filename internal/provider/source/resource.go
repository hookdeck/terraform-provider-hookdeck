package source

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/generated/tfplugingen/resource_source"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

var (
	_ resource.Resource                = &sourceResource{}
	_ resource.ResourceWithConfigure   = &sourceResource{}
	_ resource.ResourceWithImportState = &sourceResource{}
)

func NewSourceResource() resource.Resource {
	return &sourceResource{}
}

type sourceResource struct {
	client hookdeckClient.Client
}

func (r *sourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source"
}

func (r *sourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_source.SourceResourceSchema(ctx)
}

func (r *sourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hookdeckClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hookdeckClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = *client
}

func (r *sourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_source.SourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.retrieveSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_source.SourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.createSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_source.SourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_source.SourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.deleteSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *sourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sourceResource) retrieveSource(ctx context.Context, data *resource_source.SourceModel) diag.Diagnostics {
	source, err := r.client.Source.Retrieve(context.Background(), data.Id.ValueString(), &hookdeck.SourceRetrieveRequest{})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error reading source", err.Error())
		return diags
	}

	return r.refreshData(ctx, data, source)
}

func (r *sourceResource) createSource(ctx context.Context, data *resource_source.SourceModel) diag.Diagnostics {
	payload, diags := r.dataToCreatePayload(ctx, data)
	if diags.HasError() {
		return diags
	}

	source, err := r.client.Source.Create(context.Background(), payload)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	return r.refreshData(ctx, data, source)
}

func (r *sourceResource) updateSource(ctx context.Context, data *resource_source.SourceModel) diag.Diagnostics {
	payload, diags := r.dataToUpdatePayload(ctx, data)
	if diags.HasError() {
		return diags
	}

	source, err := r.client.Source.Update(context.Background(), data.Id.ValueString(), payload)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	return r.refreshData(ctx, data, source)
}

func (r *sourceResource) deleteSource(ctx context.Context, data *resource_source.SourceModel) diag.Diagnostics {
	if _, err := r.client.Source.Update(context.Background(), data.Id.ValueString(), &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error deleting source", err.Error())
		return diags
	}

	return nil
}

func (r *sourceResource) dataToCreatePayload(_ context.Context, data *resource_source.SourceModel) (*hookdeck.SourceCreateRequest, diag.Diagnostics) {
	payload := &hookdeck.SourceCreateRequest{
		Name:        data.Name.ValueString(),
		Description: hookdeck.OptionalOrNull(data.Description.ValueStringPointer()),
	}

	// TODO: config

	return payload, nil
}

func (r *sourceResource) dataToUpdatePayload(_ context.Context, data *resource_source.SourceModel) (*hookdeck.SourceUpdateRequest, diag.Diagnostics) {
	payload := &hookdeck.SourceUpdateRequest{
		Name:        hookdeck.OptionalOrNull(data.Name.ValueStringPointer()),
		Description: hookdeck.OptionalOrNull(data.Description.ValueStringPointer()),
	}

	// TODO: config

	return payload, nil
}

func (r *sourceResource) refreshData(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics {
	data.CreatedAt = types.StringValue(source.CreatedAt.Format(time.RFC3339))
	if source.DisabledAt != nil {
		data.DisabledAt = types.StringValue(source.DisabledAt.Format(time.RFC3339))
	} else {
		data.DisabledAt = types.StringNull()
	}
	data.Id = types.StringValue(source.Id)
	data.Name = types.StringValue(source.Name)
	data.TeamId = types.StringValue(source.TeamId)
	data.UpdatedAt = types.StringValue(source.UpdatedAt.Format(time.RFC3339))
	data.Url = types.StringValue(source.Url)

	// TODO: config

	return nil
}
