package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type oauth2ClientCredentialsAuthenticationMethodModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Scope        types.String `tfsdk:"scope"`
	AuthServer   types.String `tfsdk:"auth_server"`
}

type oauth2ClientCredentialsAuthenticationMethod struct {
}

func (*oauth2ClientCredentialsAuthenticationMethod) name() string {
	return "oauth2_client_credentials"
}

func (*oauth2ClientCredentialsAuthenticationMethod) schema() schema.Attribute {
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

func oauth2ClientCredentialsAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":     types.StringType,
		"client_secret": types.StringType,
		"scope":         types.StringType,
		"auth_server":   types.StringType,
	}
}

func (oauth2ClientCredentialsAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: oauth2ClientCredentialsAuthenticationMethodAttrTypesMap()}
}

func (oauth2ClientCredentialsAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(oauth2ClientCredentialsAuthenticationMethodAttrTypesMap())
}
func (oauth2ClientCredentialsAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.Oauth2ClientCredentials == nil {
		return
	}

	m.AuthMethod.OAuth2ClientCredentials = &oauth2ClientCredentialsAuthenticationMethodModel{}
	m.AuthMethod.OAuth2ClientCredentials.ClientID = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.ClientId)
	m.AuthMethod.OAuth2ClientCredentials.ClientSecret = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.ClientSecret)
	if destination.AuthMethod.Oauth2ClientCredentials.Config.Scope != nil {
		m.AuthMethod.OAuth2ClientCredentials.Scope = types.StringValue(*destination.AuthMethod.Oauth2ClientCredentials.Config.Scope)
	}
	m.AuthMethod.OAuth2ClientCredentials.AuthServer = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.AuthServer)
}

func (oauth2ClientCredentialsAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.OAuth2ClientCredentials == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromOauth2ClientCredentials(&hookdeck.AuthOAuth2ClientCredentials{
		Config: &hookdeck.DestinationAuthMethodOAuth2ClientCredentialsConfig{
			ClientId:     method.OAuth2ClientCredentials.ClientID.ValueString(),
			ClientSecret: method.OAuth2ClientCredentials.ClientSecret.ValueString(),
			Scope:        method.OAuth2ClientCredentials.Scope.ValueStringPointer(),
			AuthServer:   method.OAuth2ClientCredentials.AuthServer.ValueString(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &oauth2ClientCredentialsAuthenticationMethod{})
}
