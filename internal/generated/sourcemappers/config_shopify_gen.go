package sourcemappers

import (
	"context"
	"encoding/json"
	"terraform-provider-hookdeck/internal/generated/tfplugingen/resource_source"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type ShopifyMapper struct{}

var _ SourceTypeMapper = &ShopifyMapper{}

func init() {
	SourceTypeMappers["shopify"] = &ShopifyMapper{}
}

func (h *ShopifyMapper) Refresh(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics {
	if source.Type != "SHOPIFY" {
		return nil
	}

	configData := map[string]attr.Value{}

	// === Parse data from source ===
	// === End of parsing data from source ===

	var diags diag.Diagnostics
	data.Config.Ebay, diags = resource_source.NewEbayValueMust(resource_source.EbayValue{}.AttributeTypes(ctx), configData).ToObjectValue(ctx)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (h *ShopifyMapper) DataToSourceTypeConfig(ctx context.Context, data *resource_source.SourceModel, config *hookdeck.SourceTypeConfig) diag.Diagnostics {
	var payload *hookdeck.SourceTypeConfigHttp
	var configData map[string]interface{}
	if diags := data.Config.Http.As(ctx, &configData, basetypes.ObjectAsOptions{}); diags.HasError() {
		return diags
	}
	configDataBytes, err := json.Marshal(configData)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error marshalling config", err.Error())
		return diags
	}
	if err := payload.UnmarshalJSON(configDataBytes); err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error unmarshalling config", err.Error())
		return diags
	}
	config.SourceTypeConfigHttp = payload
	return nil
}

func (h *ShopifyMapper) DataToCreateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceCreateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceCreateRequestTypeFromString("SHOPIFY")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}

func (h *ShopifyMapper) DataToUpdateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceUpdateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("SHOPIFY")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}
