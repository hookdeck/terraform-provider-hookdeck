package sourceverification

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func awsSNSConfigSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:   true,
		Attributes: map[string]schema.Attribute{},
	}
}

type awsSNSSourceVerification struct {
}

func (m *awsSNSSourceVerification) toPayload() *hookdeck.VerificationConfig {
	return hookdeck.NewVerificationConfigFromAwsSns(&hookdeck.VerificationAwssns{
		Configs: &hookdeck.VerificationAwssnsConfigs{},
	})
}
