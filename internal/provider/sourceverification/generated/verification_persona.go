// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type personaSourceVerification struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}

type personaSourceVerificationProvider struct {
}

func (p *personaSourceVerificationProvider) getSchemaName() string {
	return "persona"
}

func (p *personaSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
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

func (p *personaSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Persona == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromPersona(&hookdeck.VerificationPersona{
		Configs: &hookdeck.VerificationPersonaConfigs{
			WebhookSecretKey: sourceVerification.Persona.WebhookSecretKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &personaSourceVerificationProvider{})
}