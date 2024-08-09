package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type SourceVerification struct {
	ApiKey *apiKeySourceVerification `tfsdk:"api_key"`
	Stripe *stripeSourceVerification `tfsdk:"stripe"`
	Github *githubSourceVerification `tfsdk:"github"`
}

type sourceVerificationProvider interface {
	getSchemaName() string
	getSchemaValue() schema.SingleNestedAttribute
	ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig
}

var Providers []sourceVerificationProvider

func GetSourceVerificationSchemaAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{}

	for _, provider := range Providers {
		attributes[provider.getSchemaName()] = provider.getSchemaValue()
	}

	return attributes
}
