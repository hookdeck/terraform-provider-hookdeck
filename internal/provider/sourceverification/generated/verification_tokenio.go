// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type tokenioSourceVerification struct {
	PublicKey types.String `tfsdk:"public_key"`
}

type tokenioSourceVerificationProvider struct {
}

func (p *tokenioSourceVerificationProvider) getSchemaName() string {
	return "tokenio"
}

func (p *tokenioSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
		},
	}
}

func (p *tokenioSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.Tokenio == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromTokenio(&hookdeck.VerificationTokenIo{
		Configs: &hookdeck.VerificationTokenIoConfigs{
			PublicKey: sourceVerification.Tokenio.PublicKey.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &tokenioSourceVerificationProvider{})
}
