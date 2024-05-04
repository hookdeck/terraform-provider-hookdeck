package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type sourceVerificationProvider interface {
	getSchemaName() string
	getSchemaValue() schema.SingleNestedAttribute
	toPayload(sourceVerification *sourceVerification) *hookdeck.VerificationConfig
}

var providers []sourceVerificationProvider

func getSourceVerificationSchemaAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{}

	for _, provider := range providers {
		attributes[provider.getSchemaName()] = provider.getSchemaValue()
	}

	return attributes
}
