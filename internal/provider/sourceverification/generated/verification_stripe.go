package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type stripeSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type stripeSourceVerificationProvider struct {
}

func (p *stripeSourceVerificationProvider) getSchemaName() string {
	return "stripe"
}

func (p *stripeSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *stripeSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Stripe == nil {
		return nil
	}

	return hookdeck.NewVerificationConfigFromStripe(&hookdeck.VerificationStripe{
		Configs: &hookdeck.VerificationStripeConfigs{
			WebhookSecretKey: sourceVerification.Stripe.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &stripeSourceVerificationProvider{})
}
