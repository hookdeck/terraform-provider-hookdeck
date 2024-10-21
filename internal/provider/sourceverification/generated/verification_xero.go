// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type xeroSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type xeroSourceVerificationProvider struct {
}

func (p *xeroSourceVerificationProvider) getSchemaName() string {
	return "xero"
}

func (p *xeroSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *xeroSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Xero == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromXero(&hookdeck.VerificationXero{
		Configs: &hookdeck.VerificationXeroConfigs{
			WebhookSecretKey: sourceVerification.Xero.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &xeroSourceVerificationProvider{})
}