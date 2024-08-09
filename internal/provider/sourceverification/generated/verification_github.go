package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type githubSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type githubSourceVerificationProvider struct {
}

func (p *githubSourceVerificationProvider) getSchemaName() string {
	return "github"
}

func (p *githubSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *githubSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Github == nil {
		return nil
	}

	return hookdeck.NewVerificationConfigFromGithub(&hookdeck.VerificationGitHub{
		Configs: &hookdeck.VerificationGitHubConfigs{
			WebhookSecretKey: sourceVerification.Github.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &githubSourceVerificationProvider{})
}
