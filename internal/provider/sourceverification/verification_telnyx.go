package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func telnyxConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type telnyxSourceVerification struct {
	PublicKey types.String `tfsdk:"public_key"`
}

func (m *telnyxSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationTelnyx(&hookdeck.VerificationTelnyx{
		Type: hookdeck.VerificationTelnyxTypeTelnyx,
		Configs: &hookdeck.VerificationTelnyxConfigs{
			PublicKey: m.PublicKey.ValueString(),
		},
	})
}
