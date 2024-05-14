package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func cloudsignalConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type cloudsignalSourceVerification struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (m *cloudsignalSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromCloudsignal(&hookdeck.VerificationCloudSignal{
		Configs: &hookdeck.VerificationCloudSignalConfigs{
			ApiKey: m.APIKey.ValueString(),
		},
	})
}
