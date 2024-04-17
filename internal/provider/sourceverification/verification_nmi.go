package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func nmiConfigSchema() schema.SingleNestedAttribute {
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

type nmiSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *nmiSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationNmiPaymentGateway(&hookdeck.VerificationNmiPaymentGateway{
		Type: hookdeck.VerificationNmiPaymentGatewayTypeNmi,
		Configs: &hookdeck.VerificationNmiPaymentGatewayConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
