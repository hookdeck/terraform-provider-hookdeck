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

type EbayMapper struct{}

var _ SourceTypeMapper = &EbayMapper{}

func init() {
	SourceTypeMappers["ebay"] = &EbayMapper{}
}

func (h *EbayMapper) Refresh(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics {
	if source.Type != "EBAY" {
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

func (h *EbayMapper) DataToSourceTypeConfig(ctx context.Context, data *resource_source.SourceModel, config *hookdeck.SourceTypeConfig) diag.Diagnostics {
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

func (h *EbayMapper) DataToCreateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceCreateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceCreateRequestTypeFromString("EBAY")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}

func (h *EbayMapper) DataToUpdateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceUpdateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("EBAY")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}
