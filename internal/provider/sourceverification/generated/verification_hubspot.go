// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type hubspotSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type hubspotSourceVerificationProvider struct {
}

func (p *hubspotSourceVerificationProvider) getSchemaName() string {
	return "hubspot"
}

func (p *hubspotSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *hubspotSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Hubspot == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromHubspot(&hookdeck.VerificationHubspot{
		Configs: &hookdeck.VerificationHubspotConfigs{
			WebhookSecretKey: sourceVerification.Hubspot.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &hubspotSourceVerificationProvider{})
}