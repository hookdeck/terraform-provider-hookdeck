package sourcemappers

import (
	"context"
	"encoding/json"
	"terraform-provider-hookdeck/internal/generated/tfplugingen/resource_source"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

var SourceTypeMappers = map[string]SourceTypeMapper{}

type SourceTypeMapper interface {
	Refresh(ctx context.Context, data *resource_source.SourceModel, source *hookdeck.Source) diag.Diagnostics
	DataToSourceTypeConfig(ctx context.Context, data *resource_source.SourceModel, config *hookdeck.SourceTypeConfig) diag.Diagnostics
	DataToCreateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceCreateRequestType, diag.Diagnostics)
	DataToUpdateSourceType(ctx context.Context, data *resource_source.SourceModel) (*hookdeck.SourceUpdateRequestType, diag.Diagnostics)
}

func getConfig(source *hookdeck.Source) map[string]interface{} {
	sourceBytes, _ := source.MarshalJSON()
	var sourceData map[string]interface{}
	if err := json.Unmarshal(sourceBytes, &sourceData); err != nil {
		panic(err)
	}
	return sourceData["config"].(map[string]interface{})
}
