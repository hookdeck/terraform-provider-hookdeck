// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type workosSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type workosSourceVerificationProvider struct {
}

func (p *workosSourceVerificationProvider) getSchemaName() string {
	return "workos"
}

func (p *workosSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *workosSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Workos == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromWorkos(&hookdeck.VerificationWorkOs{
		Configs: &hookdeck.VerificationWorkOsConfigs{
			WebhookSecretKey: sourceVerification.Workos.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &workosSourceVerificationProvider{})
}
