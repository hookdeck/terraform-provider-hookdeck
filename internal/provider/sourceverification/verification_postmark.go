package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func postmarkConfigSchema() schema.SingleNestedAttribute {
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

type postmarkSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *postmarkSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationPostmark(&hookdeck.VerificationPostmark{
		Configs: &hookdeck.VerificationPostmarkConfigs{
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
