package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type apiKeySourceVerification struct {
	HeaderKey types.String `tfsdk:"header_key"`
	ApiKey    types.String `tfsdk:"api_key"`
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
			"header_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *apiKeySourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.ApiKey == nil {
		return nil
	}

	return hookdeck.NewVerificationConfigFromApiKey(&hookdeck.VerificationApiKey{
		Configs: &hookdeck.VerificationApiKeyConfigs{
			HeaderKey: sourceVerification.ApiKey.HeaderKey.ValueString(),
			ApiKey:    sourceVerification.ApiKey.ApiKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &apiKeySourceVerificationProvider{})
}
