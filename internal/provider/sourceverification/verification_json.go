package sourceverification

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func jsonConfigSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
	}
}

func jsonToPayload(stringifiedJSON string) *hookdeck.VerificationConfig {
	var verification *hookdeck.VerificationConfig
	json.Unmarshal([]byte(stringifiedJSON), &verification)
	return verification
}
