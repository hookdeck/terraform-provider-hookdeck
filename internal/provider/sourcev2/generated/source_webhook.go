package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"terraform-provider-hookdeck/internal/generated/resource_source_config_webhook"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// ============================================================================
// Source Config Resource
// ============================================================================

// Source Config Resource
var (
	_ resource.Resource                = &webhookSourceConfigResource{}
	_ resource.ResourceWithConfigure   = &webhookSourceConfigResource{}
	_ resource.ResourceWithImportState = &webhookSourceConfigResource{}
)

func init() {
	newResources = append(newResources, NewWebhookSourceConfigResource)
}

func NewWebhookSourceConfigResource() resource.Resource {
	return &webhookSourceConfigResource{}
}

type webhookSourceConfigResource struct {
	client hookdeckClient.Client
}

func (r *webhookSourceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_source_config_webhook"
}

func (r *webhookSourceConfigResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_source_config_webhook.SourceConfigWebhookResourceSchema(ctx)
}

func (r *webhookSourceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookSourceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_source_config_webhook.SourceConfigWebhookModel
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

func (r *webhookSourceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_source_config_webhook.SourceConfigWebhookModel

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

func (r *webhookSourceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_source_config_webhook.SourceConfigWebhookModel

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

func (r *webhookSourceConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_source_config_webhook.SourceConfigWebhookModel

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

func (r *webhookSourceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webhookSourceConfigResource) readSourceConfig(ctx context.Context, data *resource_source_config_webhook.SourceConfigWebhookModel) diag.Diagnostics {
	source, err := r.client.Source.Retrieve(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error reading source config", err.Error())
		return diags
	}

	return r.refreshData(ctx, data, source)
}

func (r *webhookSourceConfigResource) updateSourceConfig(ctx context.Context, data *resource_source_config_webhook.SourceConfigWebhookModel) diag.Diagnostics {
	payload, diags := r.dataToUpdatePayload(data)
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

	return r.refreshData(ctx, data, source)
}

func (r *webhookSourceConfigResource) deleteSourceConfig(ctx context.Context, data *resource_source_config_webhook.SourceConfigWebhookModel) diag.Diagnostics {
	if _, err := r.client.Source.Update(context.Background(), data.SourceId.ValueString(), &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error deleting source config", err.Error())
		return diags
	}

	return nil
}

func (r *webhookSourceConfigResource) dataToUpdatePayload(data *resource_source_config_webhook.SourceConfigWebhookModel) (*hookdeck.SourceUpdateRequest, diag.Diagnostics) {
	config := &hookdeck.SourceTypeConfigWebhook{}

	if !data.Auth.IsUnknown() && !data.Auth.IsNull() {
		object, diags := data.Auth.ToObjectValue(context.Background())
		if diags.HasError() {
			return nil, diags
		}
		var auth hookdeck.SourceTypeConfigWebhookAuth
		if err := json.Unmarshal([]byte(object.String()), &auth); err != nil {
			var diags diag.Diagnostics
			diags.AddError("Error unmarshalling auth", err.Error())
			return nil, diags
		}
		config.Auth = &auth
	}

	if !data.AllowedHttpMethods.IsUnknown() && !data.AllowedHttpMethods.IsNull() {
		for _, method := range data.AllowedHttpMethods.Elements() {
			method, err := hookdeck.NewSourceTypeConfigWebhookAllowedHttpMethodsItemFromString(method.(types.String).ValueString())
			if err != nil {
				var diags diag.Diagnostics
				diags.AddError("Error converting allowed http method to SourceTypeConfigWebhookAllowedHttpMethodsItem", err.Error())
				return nil, diags
			}
			config.AllowedHttpMethods = append(config.AllowedHttpMethods, method)
		}
	}

	if !data.CustomResponse.IsUnknown() && !data.CustomResponse.IsNull() {
		object, diags := data.CustomResponse.ToObjectValue(context.Background())
		if diags.HasError() {
			return nil, diags
		}
		var apiObject hookdeck.SourceTypeConfigWebhookCustomResponse
		if err := json.Unmarshal([]byte(object.String()), &apiObject); err != nil {
			var diags diag.Diagnostics
			diags.AddError("Error unmarshalling custom_response", err.Error())
			return nil, diags
		}
		config.CustomResponse = &apiObject
	}

	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("WEBHOOK")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error converting source type to SourceUpdateRequestType", err.Error())
		return nil, diags
	}

	return &hookdeck.SourceUpdateRequest{
		Type: hookdeck.Optional(sourceType),
		Config: hookdeck.Optional(hookdeck.SourceTypeConfig{
			SourceTypeConfigWebhook: config,
		}),
	}, nil
}

func (r *webhookSourceConfigResource) refreshData(ctx context.Context, data *resource_source_config_webhook.SourceConfigWebhookModel, source *hookdeck.Source) diag.Diagnostics {
	var diags diag.Diagnostics

	if source.AllowedHttpMethods != nil {
		data.AllowedHttpMethods, diags = types.ListValueFrom(ctx, types.StringType, source.AllowedHttpMethods)
		if diags.HasError() {
			return diags
		}
	}

	if source.Verification != nil {
		sourceData, err := toSourceJSON(source)
		if err != nil {
			diags.AddError("Error converting source to JSON", err.Error())
			return diags
		}
		authConfig := getAuthConfig(sourceData)
		data.Type = types.StringValue(authConfig["type"].(string))
	}

	if source.CustomResponse != nil {
		data.CustomResponse.Body = types.StringValue(source.CustomResponse.Body)
		data.CustomResponse.ContentType = types.StringValue(string(source.CustomResponse.ContentType))
	}

	return nil
}
