package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func apiKeyConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"header_key": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

type apiKeySourceVerification struct {
	APIKey    types.String `tfsdk:"api_key"`
	HeaderKey types.String `tfsdk:"header_key"`
}

func (m *apiKeySourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationApiKey(&hookdeck.VerificationApiKey{
		Type: hookdeck.VerificationApiKeyTypeApiKey,
		Configs: &hookdeck.VerificationApiKeyConfigs{
			ApiKey:    m.APIKey.ValueString(),
			HeaderKey: m.HeaderKey.ValueString(),
		},
	})
}
