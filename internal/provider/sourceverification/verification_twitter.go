package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func twitterConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type twitterSourceVerification struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (m *twitterSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationTwitter(&hookdeck.VerificationTwitter{
		Configs: &hookdeck.VerificationTwitterConfigs{
			ApiKey: m.APIKey.ValueString(),
		},
	})
}
