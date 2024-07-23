package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func vercelLogDrainsConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"webhook_secret_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"vercel_log_drains_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type vercelLogDrainsSourceVerification struct {
	WebhookSecretKey      types.String `tfsdk:"webhook_secret_key"`
	VercelLogDrainsSecret types.String `tfsdk:"vercel_log_drains_secret"`
}

func (m *vercelLogDrainsSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVercelLogDrains(&hookdeck.VerificationVercelLogDrains{
		Configs: &hookdeck.VerificationVercelLogDrainsConfigs{
			WebhookSecretKey:      m.WebhookSecretKey.ValueStringPointer(),
			VercelLogDrainsSecret: m.VercelLogDrainsSecret.ValueString(),
		},
	})
}
