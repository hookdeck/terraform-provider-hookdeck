// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type orbSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type orbSourceVerificationProvider struct {
}

func (p *orbSourceVerificationProvider) getSchemaName() string {
	return "orb"
}

func (p *orbSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *orbSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Orb == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromOrb(&hookdeck.VerificationOrb{
		Configs: &hookdeck.VerificationOrbConfigs{
			WebhookSecretKey: sourceVerification.Orb.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &orbSourceVerificationProvider{})
}
