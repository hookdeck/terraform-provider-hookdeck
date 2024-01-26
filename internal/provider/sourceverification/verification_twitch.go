package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func twitchConfigSchema() schema.SingleNestedAttribute {
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

type twitchSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *twitchSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromTwitch(&hookdeck.VerificationTwitch{
		Configs: &hookdeck.VerificationTwitchConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
