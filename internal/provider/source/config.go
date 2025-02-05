package source

import (
	"context"
	"terraform-provider-hookdeck/internal/generated/sourcemappers"
	"terraform-provider-hookdeck/internal/generated/tfplugingen/resource_source"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func refreshSourceConfig(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics {
	nullConfigValue, diags := resource_source.NewConfigValueNull().ToObjectValue(ctx)
	if diags.HasError() {
		return diags
	}
	data.Config, diags = resource_source.NewConfigValue(resource_source.ConfigValue{}.AttributeTypes(ctx), nullConfigValue.Attributes())
	if diags.HasError() {
		return diags
	}

	for _, mapper := range sourcemappers.SourceTypeMappers {
		if diags := mapper.Refresh(ctx, data, source); diags.HasError() {
			return diags
		}
	}

	return nil
}

func configCreatePayload(ctx context.Context, data *resource_source.SourceModel, payload *hookdeck.SourceCreateRequest) diag.Diagnostics {
	config := &hookdeck.SourceTypeConfig{}
	var sourceType *hookdeck.SourceCreateRequestType

	for _, mapper := range sourcemappers.SourceTypeMappers {
		var diags diag.Diagnostics
		sourceType, diags = mapper.DataToCreateSourceType(ctx, data)
		if diags.HasError() {
			return diags
		}
		if sourceType != nil {
			if diags := mapper.DataToSourceTypeConfig(ctx, data, config); diags.HasError() {
				return diags
			}
			break
		}
	}

	payload.Type = hookdeck.OptionalOrNull(sourceType)
	payload.Config = hookdeck.OptionalOrNull(config)

	return nil
}

func configUpdatePayload(ctx context.Context, data *resource_source.SourceModel, payload *hookdeck.SourceUpdateRequest) diag.Diagnostics {
	config := &hookdeck.SourceTypeConfig{}
	var sourceType *hookdeck.SourceUpdateRequestType

	for _, mapper := range sourcemappers.SourceTypeMappers {
		var diags diag.Diagnostics
		sourceType, diags = mapper.DataToUpdateSourceType(ctx, data)
		if diags.HasError() {
			return diags
		}
		if sourceType != nil {
			if diags := mapper.DataToSourceTypeConfig(ctx, data, config); diags.HasError() {
				return diags
			}
			break
		}
	}

	payload.Type = hookdeck.OptionalOrNull(sourceType)
	payload.Config = hookdeck.OptionalOrNull(config)

	return nil
}
