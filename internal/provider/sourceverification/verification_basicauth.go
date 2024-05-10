package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func basicAuthConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

type basicAuthSourceVerification struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (m *basicAuthSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationBasicAuth(&hookdeck.VerificationBasicAuth{
		Type: hookdeck.VerificationBasicAuthTypeBasicAuth,
		Configs: &hookdeck.VerificationBasicAuthConfigs{
			Username: m.Username.ValueString(),
			Password: m.Password.ValueString(),
		},
	})
}
