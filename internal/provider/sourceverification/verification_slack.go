package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func slackConfigSchema() schema.SingleNestedAttribute {
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

type slackSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *slackSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromSlack(&hookdeck.VerificationSlack{
		Configs: &hookdeck.VerificationSlackConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
