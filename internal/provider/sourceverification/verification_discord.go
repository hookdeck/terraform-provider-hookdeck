package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func discordConfigSchema() schema.SingleNestedAttribute {
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

type discordSourceVerification struct {
	PublicKey types.String `tfsdk:"public_key"`
}

func (m *discordSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromDiscord(&hookdeck.VerificationDiscord{
		Configs: &hookdeck.VerificationDiscordConfigs{
			PublicKey: m.PublicKey.ValueString(),
		},
	})
}
