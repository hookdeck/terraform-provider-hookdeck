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
			"name": schema.StringAttribute{
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
	Name     types.String `tfsdk:"name"`
	Password types.String `tfsdk:"password"`
}

func (m *basicAuthSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationBasicAuth(&hookdeck.VerificationBasicAuth{
		Configs: &hookdeck.VerificationBasicAuthConfigs{
			Name:     m.Name.ValueString(),
			Password: m.Password.ValueString(),
		},
	})
}
