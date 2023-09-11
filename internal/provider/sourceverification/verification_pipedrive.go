package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func pipedriveConfigSchema() schema.SingleNestedAttribute {
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

type pipedriveSourceVerification struct {
	Name     types.String `tfsdk:"name"`
	Password types.String `tfsdk:"password"`
}

func (m *pipedriveSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationPipedrive(&hookdeck.VerificationPipedrive{
		Configs: &hookdeck.VerificationPipedriveConfigs{
			Name:     m.Name.ValueString(),
			Password: m.Password.ValueString(),
		},
	})
}
