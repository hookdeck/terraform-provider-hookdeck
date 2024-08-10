// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type fiservSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type fiservSourceVerificationProvider struct {
}

func (p *fiservSourceVerificationProvider) getSchemaName() string {
	return "fiserv"
}

func (p *fiservSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *fiservSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Fiserv == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromFiserv(&hookdeck.VerificationFiserv{
		Configs: &hookdeck.VerificationFiservConfigs{
			WebhookSecretKey: sourceVerification.Fiserv.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &fiservSourceVerificationProvider{})
}
