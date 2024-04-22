package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func rechargeConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"algorithm": schema.StringAttribute{
				// TODO: enum "md5, sha1, sha256, sha512"
				Required: true,
			},
			"encoding": schema.StringAttribute{
				// TODO: enum "base64, hex"
				Required: true,
			},
			"header_key": schema.StringAttribute{
				Required: true,
			},
			"webhook_secret_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type rechargeSourceVerification struct {
	Algorithm        types.String `tfsdk:"algorithm"`
	Encoding         types.String `tfsdk:"encoding"`
	HeaderKey        types.String `tfsdk:"header_key"`
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

func (m *rechargeSourceVerification) toPayload() *hookdeck.VerificationConfig {
	algorithm, _ := hookdeck.NewHmacAlgorithmsFromString(m.Algorithm.ValueString())
	encoding, _ := hookdeck.NewVerificationRechargeConfigsEncodingFromString(m.Encoding.ValueString())
	return hookdeck.NewVerificationConfigFromVerificationRecharge(&hookdeck.VerificationRecharge{
		Type: hookdeck.VerificationRechargeTypeRecharge,
		Configs: &hookdeck.VerificationRechargeConfigs{
			Algorithm:        algorithm,
			Encoding:         encoding,
			HeaderKey:        m.HeaderKey.ValueString(),
			WebhookSecretKey: m.WebhookSecretKey.ValueString(),
		},
	})
}
