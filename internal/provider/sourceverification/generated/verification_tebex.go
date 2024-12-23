// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type tebexSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type tebexSourceVerificationProvider struct {
}

func (p *tebexSourceVerificationProvider) getSchemaName() string {
	return "tebex"
}

func (p *tebexSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *tebexSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Tebex == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromTebex(&hookdeck.VerificationTebex{
		Configs: &hookdeck.VerificationTebexConfigs{
			WebhookSecretKey: sourceVerification.Tebex.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &tebexSourceVerificationProvider{})
}
