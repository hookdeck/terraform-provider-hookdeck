// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type adyenSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type adyenSourceVerificationProvider struct {
}

func (p *adyenSourceVerificationProvider) getSchemaName() string {
	return "adyen"
}

func (p *adyenSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"webhook_secret_key": schema.StringAttribute{
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
		},
	}
}

func (p *adyenSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Adyen == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromAdyen(&hookdeck.VerificationAdyen{
		Configs: &hookdeck.VerificationAdyenConfigs{
			WebhookSecretKey: sourceVerification.Adyen.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &adyenSourceVerificationProvider{})
}