// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type basicAuthSourceVerification struct {
	Password types.String `tfsdk:"password"`
	Username types.String `tfsdk:"username"`
}

type basicAuthSourceVerificationProvider struct {
}

func (p *basicAuthSourceVerificationProvider) getSchemaName() string {
	return "basic_auth"
}

func (p *basicAuthSourceVerificationProvider) getSchemaValue() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				Required:  true,
				Optional:  false,
				Sensitive: true,
			},
		},
	}
}

func (p *basicAuthSourceVerificationProvider) ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig {
	if sourceVerification.BasicAuth == nil {
		return nil
	}
	return hookdeck.NewVerificationConfigFromBasicAuth(&hookdeck.VerificationBasicAuth{
		Configs: &hookdeck.VerificationBasicAuthConfigs{
			Password: sourceVerification.BasicAuth.Password.ValueString(),
			Username: sourceVerification.BasicAuth.Username.ValueString(),
		},
	})
}

func init() {
	Providers = append(Providers, &basicAuthSourceVerificationProvider{})
}