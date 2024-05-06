package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type oauth2AuthorizationCodeAuthenticationMethodModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	RefreshToken types.String `tfsdk:"refresh_token"`
	Scope        types.String `tfsdk:"scope"`
	AuthServer   types.String `tfsdk:"auth_server"`
}

type oauth2AuthorizationCodeAuthenticationMethod struct {
}

func (*oauth2AuthorizationCodeAuthenticationMethod) name() string {
	return "oauth2_authorization_code"
}

func (*oauth2AuthorizationCodeAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: `Client id in the auth server`,
			},
			"client_secret": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `Client secret in the auth server`,
			},
			"refresh_token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `Refresh token already returned by the auth server`,
			},
			"scope": schema.StringAttribute{
				Optional:    true,
				Description: `Scope to access`,
			},
			"auth_server": schema.StringAttribute{
				Required:    true,
				Description: `URL of the auth server`,
			},
		},
		Description: `OAuth2 Client Credentials`,
	}
}

func oauth2AuthorizationCodeAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":     types.StringType,
		"client_secret": types.StringType,
		"refresh_token": types.StringType,
		"scope":         types.StringType,
		"auth_server":   types.StringType,
	}
}

func (oauth2AuthorizationCodeAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: oauth2AuthorizationCodeAuthenticationMethodAttrTypesMap()}
}

func (oauth2AuthorizationCodeAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(oauth2AuthorizationCodeAuthenticationMethodAttrTypesMap())
}

func (oauth2AuthorizationCodeAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.Oauth2AuthorizationCode == nil {
		return
	}

	m.AuthMethod.OAuth2AuthorizationCode = &oauth2AuthorizationCodeAuthenticationMethodModel{}
	m.AuthMethod.OAuth2AuthorizationCode.ClientID = types.StringValue(destination.AuthMethod.Oauth2AuthorizationCode.Config.ClientId)
	m.AuthMethod.OAuth2AuthorizationCode.ClientSecret = types.StringValue(destination.AuthMethod.Oauth2AuthorizationCode.Config.ClientSecret)
	m.AuthMethod.OAuth2AuthorizationCode.RefreshToken = types.StringValue(destination.AuthMethod.Oauth2AuthorizationCode.Config.RefreshToken)
	if destination.AuthMethod.Oauth2AuthorizationCode.Config.Scope != nil {
		m.AuthMethod.OAuth2AuthorizationCode.Scope = types.StringValue(*destination.AuthMethod.Oauth2AuthorizationCode.Config.Scope)
	}
	m.AuthMethod.OAuth2AuthorizationCode.AuthServer = types.StringValue(destination.AuthMethod.Oauth2AuthorizationCode.Config.AuthServer)
}

func (oauth2AuthorizationCodeAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.OAuth2AuthorizationCode == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromOauth2AuthorizationCode(&hookdeck.AuthOAuth2AuthorizationCode{
		Config: &hookdeck.DestinationAuthMethodOAuth2AuthorizationCodeConfig{
			ClientId:     method.OAuth2AuthorizationCode.ClientID.ValueString(),
			ClientSecret: method.OAuth2AuthorizationCode.ClientSecret.ValueString(),
			RefreshToken: method.OAuth2AuthorizationCode.RefreshToken.ValueString(),
			Scope:        method.OAuth2AuthorizationCode.Scope.ValueStringPointer(),
			AuthServer:   method.OAuth2AuthorizationCode.AuthServer.ValueString(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &oauth2AuthorizationCodeAuthenticationMethod{})
}
