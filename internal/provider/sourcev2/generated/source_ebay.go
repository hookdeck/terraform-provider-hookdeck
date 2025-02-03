package generated

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/generated/resource_source_config_ebay"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// ============================================================================
// Source Config Resource
// ============================================================================

// Source Config Resource
var (
	_ resource.Resource                = &ebaySourceConfigResource{}
	_ resource.ResourceWithConfigure   = &ebaySourceConfigResource{}
	_ resource.ResourceWithImportState = &ebaySourceConfigResource{}
)

func init() {
	newResources = append(newResources, NewEbaySourceConfigResource)
}

func NewEbaySourceConfigResource() resource.Resource {
	return &ebaySourceConfigResource{}
}

type ebaySourceConfigResource struct {
	client hookdeckClient.Client
}

func (r *ebaySourceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ebay_source_config"
}

func (r *ebaySourceConfigResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_source_config_ebay.SourceConfigEbayResourceSchema(ctx)
}

func (r *ebaySourceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ebaySourceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_source_config_ebay.SourceConfigEbayModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readSourceConfig(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_source_config_ebay.SourceConfigEbayModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateSourceConfig(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_source_config_ebay.SourceConfigEbayModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.updateSourceConfig(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_source_config_ebay.SourceConfigEbayModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Update(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source config", err.Error())
		return
	}
}

func (r *ebaySourceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ebaySourceConfigResource) readSourceConfig(ctx context.Context, data *resource_source_config_ebay.SourceConfigEbayModel) diag.Diagnostics {
	source, err := r.client.Source.Retrieve(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error reading source config", err.Error())
		return diags
	}

	return refreshData(data, source)
}

func (r *ebaySourceConfigResource) updateSourceConfig(ctx context.Context, data *resource_source_config_ebay.SourceConfigEbayModel) diag.Diagnostics {
	payload, diags := dataToUpdatePayload(data)
	if diags.HasError() {
		return diags
	}

	_, err := r.client.Source.Update(context.Background(), data.SourceId.ValueString(), payload)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source config", err.Error())
		return diags
	}

	source, err := r.client.Source.Retrieve(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error reading source config", err.Error())
		return diags
	}

	return refreshData(data, source)
}

func (r *ebaySourceConfigResource) deleteSourceConfig(ctx context.Context, data *resource_source_config_ebay.SourceConfigEbayModel) diag.Diagnostics {
	if _, err := r.client.Source.Update(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error deleting source config", err.Error())
		return diags
	}

	return nil
}

func dataToUpdatePayload(data *resource_source_config_ebay.SourceConfigEbayModel) (*hookdeck.SourceUpdateRequest, diag.Diagnostics) {
	// TODO
	return &hookdeck.SourceUpdateRequest{}, nil
}

func refreshData(data *resource_source_config_ebay.SourceConfigEbayModel, source *hookdeck.Source) diag.Diagnostics {
	// TODO
	return nil
}
