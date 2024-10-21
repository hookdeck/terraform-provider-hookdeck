// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type bondsmithSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type bondsmithSourceVerificationProvider struct {
}

func (p *bondsmithSourceVerificationProvider) getSchemaName() string {
	return "bondsmith"
}

func (p *bondsmithSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *bondsmithSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Bondsmith == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromBondsmith(&hookdeck.VerificationBondsmith{
		Configs: &hookdeck.VerificationBondsmithConfigs{
			WebhookSecretKey: sourceVerification.Bondsmith.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &bondsmithSourceVerificationProvider{})
}