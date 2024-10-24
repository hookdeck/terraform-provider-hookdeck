// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type commercelayerSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type commercelayerSourceVerificationProvider struct {
}

func (p *commercelayerSourceVerificationProvider) getSchemaName() string {
	return "commercelayer"
}

func (p *commercelayerSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *commercelayerSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Commercelayer == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromCommercelayer(&hookdeck.VerificationCommercelayer{
		Configs: &hookdeck.VerificationCommercelayerConfigs{
			WebhookSecretKey: sourceVerification.Commercelayer.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &commercelayerSourceVerificationProvider{})
}
