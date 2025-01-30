package generated

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

var newResources = []func() resource.Resource{}

func ResourceFactories() []func() resource.Resource {
	return newResources
}

// ============================================================================
// Helpers
// ============================================================================

func toSourceJSON(source *hookdeck.Source) (map[string]any, error) {
	str := source.String()
	var data map[string]any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func getAuthConfig(data map[string]any) map[string]any {
	return data["verification"].(map[string]any)["configs"].(map[string]any)
} 