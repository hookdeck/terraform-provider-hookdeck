// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type propertyFinderSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type propertyFinderSourceVerificationProvider struct {
}

func (p *propertyFinderSourceVerificationProvider) getSchemaName() string {
	return "property_finder"
}

func (p *propertyFinderSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *propertyFinderSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.PropertyFinder == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromPropertyFinder(&hookdeck.VerificationPropertyFinder{
		Configs: &hookdeck.VerificationPropertyFinderConfigs{
			WebhookSecretKey: sourceVerification.PropertyFinder.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &propertyFinderSourceVerificationProvider{})
}
