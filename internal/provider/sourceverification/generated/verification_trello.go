// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type trelloSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type trelloSourceVerificationProvider struct {
}

func (p *trelloSourceVerificationProvider) getSchemaName() string {
	return "trello"
}

func (p *trelloSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *trelloSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Trello == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromTrello(&hookdeck.VerificationTrello{
		Configs: &hookdeck.VerificationTrelloConfigs{
			WebhookSecretKey: sourceVerification.Trello.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &trelloSourceVerificationProvider{})
}
