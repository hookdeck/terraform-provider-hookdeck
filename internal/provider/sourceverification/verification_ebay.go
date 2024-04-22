package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func ebayConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"client_secret": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"dev_id": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"environment": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"verification_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type ebaySourceVerification struct {
	ClientId          types.String `tfsdk:"client_id"`
	ClientSecret      types.String `tfsdk:"client_secret"`
	DevId             types.String `tfsdk:"dev_id"`
	Environment       types.String `tfsdk:"environment"`
	VerificationToken types.String `tfsdk:"verification_token"`
}

func (m *ebaySourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationEbay(&hookdeck.VerificationEbay{
		Type: hookdeck.VerificationEbayTypeEbay,
		Configs: &hookdeck.VerificationEbayConfigs{
			ClientId:          m.ClientId.ValueString(),
			ClientSecret:      m.ClientSecret.ValueString(),
			DevId:             m.DevId.ValueString(),
			Environment:       m.Environment.ValueString(),
			VerificationToken: m.VerificationToken.ValueString(),
		},
	})
}
