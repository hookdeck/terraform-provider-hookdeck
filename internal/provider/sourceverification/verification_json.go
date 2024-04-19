package sourceverification

import (
	"encoding/json"
	"log"

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
	if err := json.Unmarshal([]byte(stringifiedJSON), &verification); err != nil {
		// TODO: improve error handling?
		log.Fatal("Error unmarshalling JSON source verification payload")
	}
	return verification
}
