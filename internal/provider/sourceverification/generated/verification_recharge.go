// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type rechargeSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type rechargeSourceVerificationProvider struct {
}

func (p *rechargeSourceVerificationProvider) getSchemaName() string {
	return "recharge"
}

func (p *rechargeSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *rechargeSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Recharge == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromRecharge(&hookdeck.VerificationRecharge{
		Configs: &hookdeck.VerificationRechargeConfigs{
			WebhookSecretKey: sourceVerification.Recharge.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &rechargeSourceVerificationProvider{})
}