package sourcemappers

import (
	"context"
	"encoding/json"
	"terraform-provider-hookdeck/internal/generated/tfplugingen/resource_source"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type HttpMapper struct{}

var _ SourceTypeMapper = &HttpMapper{}

func init() {
	SourceTypeMappers["http"] = &HttpMapper{}
}

func (h *HttpMapper) Refresh(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics {
	if source.Type != "HTTP" {
		return nil
	}

	configData := map[string]attr.Value{}

	// === Parse data from source ===
	config := getConfig(source)
	allowedHttpMethodStrings, ok := config["allowed_http_methods"].([]interface{})
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("allowed_http_methods is not a list", "allowed_http_methods is not a list")
		return diags
	}
	allowedHttpMethods := []attr.Value{}
	for _, header := range allowedHttpMethodStrings {
		allowedHttpMethods = append(allowedHttpMethods, types.StringValue(header.(string)))
	}
	configData["allowed_http_methods"] = types.ListValueMust(types.StringType, allowedHttpMethods)
	customResponse := map[string]attr.Value{}
	customResponseData, ok := config["custom_response"].(map[string]interface{})
	if !ok {
		configData["custom_response"] = types.ObjectNull(resource_source.CustomResponseValue{}.AttributeTypes(ctx))
	} else {
		customResponse["body"] = types.StringValue(customResponseData["body"].(string))
		customResponse["content_type"] = types.StringValue(customResponseData["content_type"].(string))
		configData["custom_response"] = types.ObjectValueMust(resource_source.CustomResponseValue{}.AttributeTypes(ctx), customResponse)
	}
	// === End of parsing data from source ===

	var diags diag.Diagnostics
	data.Config.Http, diags = resource_source.NewHttpValueMust(resource_source.HttpValue{}.AttributeTypes(ctx), configData).ToObjectValue(ctx)
	if diags.HasError() {
		return diags
	}
	return nil
}

func (h *HttpMapper) DataToSourceTypeConfig(ctx context.Context, data *resource_source.SourceModel, config *hookdeck.SourceTypeConfig) diag.Diagnostics {
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

func (h *HttpMapper) DataToCreateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceCreateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceCreateRequestTypeFromString("HTTP")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}

func (h *HttpMapper) DataToUpdateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceUpdateRequestType, diag.Diagnostics) {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("HTTP")
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Error creating source type", err.Error())
		return nil, diags
	}
	return &sourceType, nil
}
