package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func gitlabConfigSchema() schema.SingleNestedAttribute {
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

type gitlabSourceVerification struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (m *gitlabSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromVerificationGitLab(&hookdeck.VerificationGitLab{
		Configs: &hookdeck.VerificationGitLabConfigs{
			ApiKey: m.APIKey.ValueString(),
		},
	})
}
