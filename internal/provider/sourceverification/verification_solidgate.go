package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func solidgateConfigSchema() schema.SingleNestedAttribute {
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

type solidgateSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *solidgateSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromSolidgate(&hookdeck.VerificationSolidGate{
		Configs: &hookdeck.VerificationSolidGateConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
