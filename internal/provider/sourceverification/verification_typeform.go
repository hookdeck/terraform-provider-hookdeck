package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func typeformConfigSchema() schema.SingleNestedAttribute {
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

type typeformSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *typeformSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromTypeform(&hookdeck.VerificationTypeform{
		Configs: &hookdeck.VerificationTypeformConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
