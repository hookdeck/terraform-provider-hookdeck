package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type apiKeySourceVerification struct {
	APIKey    types.String `tfsdk:"api_key"`
	HeaderKey types.String `tfsdk:"header_key"`
}

type apiKeySourceVerificationProvider struct {
}

func (p *apiKeySourceVerificationProvider) getSchemaName() string {
	return "api_key"
}

func (p *apiKeySourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *apiKeySourceVerificationProvider) toPayload(sourceVerification *sourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.APIKey == nil {
		return nil
	}

	return hookdeck.NewVerificationConfigFromVerificationApiKey(&hookdeck.VerificationApiKey{
		Type: hookdeck.VerificationApiKeyTypeApiKey,
		Configs: &hookdeck.VerificationApiKeyConfigs{
			ApiKey:    sourceVerification.APIKey.APIKey.ValueString(),
			HeaderKey: sourceVerification.APIKey.HeaderKey.ValueString(),
		},
	})
}

func init() {
	providers = append(providers, &apiKeySourceVerificationProvider{})
}
