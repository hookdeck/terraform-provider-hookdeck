// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type paddleSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type paddleSourceVerificationProvider struct {
}

func (p *paddleSourceVerificationProvider) getSchemaName() string {
	return "paddle"
}

func (p *paddleSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *paddleSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Paddle == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromPaddle(&hookdeck.VerificationPaddle{
		Configs: &hookdeck.VerificationPaddleConfigs{
			WebhookSecretKey: sourceVerification.Paddle.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &paddleSourceVerificationProvider{})
}
