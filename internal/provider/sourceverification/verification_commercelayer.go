package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func commercelayerConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"webhook_secret_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type commercelayerSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *commercelayerSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationCommercelayer(&hookdeck.VerificationCommercelayer{
		Type: hookdeck.VerificationCommercelayerTypeCommercelayer,
		Configs: &hookdeck.VerificationCommercelayerConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
