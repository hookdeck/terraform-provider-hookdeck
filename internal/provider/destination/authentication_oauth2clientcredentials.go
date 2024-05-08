package destination

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type oauth2ClientCredentialsAuthenticationMethodModel struct {
	AuthServer         types.String `tfsdk:"auth_server"`
	AuthenticationType types.String `tfsdk:"authentication_type"`
	ClientID           types.String `tfsdk:"client_id"`
	ClientSecret       types.String `tfsdk:"client_secret"`
	Scope              types.String `tfsdk:"scope"`
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
			"auth_server": schema.StringAttribute{
				Required:    true,
				Description: `URL of the auth server`,
			},
			"authentication_type": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "must be one of [basic, bearer]" + "\n" + `Basic (default) or Bearer Authentication`,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"basic",
						"bearer",
					),
				},
				Default: stringdefault.StaticString("basic"),
			},
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
		},
		Description: `OAuth2 Client Credentials`,
	}
}

func oauth2ClientCredentialsAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_server":         types.StringType,
		"authentication_type": types.StringType,
		"client_id":           types.StringType,
		"client_secret":       types.StringType,
		"scope":               types.StringType,
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
	m.AuthMethod.OAuth2ClientCredentials.AuthServer = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.AuthServer)
	if destination.AuthMethod.Oauth2ClientCredentials.Config.AuthenticationType != nil {
		m.AuthMethod.OAuth2ClientCredentials.AuthenticationType = types.StringValue(string(*destination.AuthMethod.Oauth2ClientCredentials.Config.AuthenticationType))
	}
	m.AuthMethod.OAuth2ClientCredentials.ClientID = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.ClientId)
	m.AuthMethod.OAuth2ClientCredentials.ClientSecret = types.StringValue(destination.AuthMethod.Oauth2ClientCredentials.Config.ClientSecret)
	if destination.AuthMethod.Oauth2ClientCredentials.Config.Scope != nil {
		m.AuthMethod.OAuth2ClientCredentials.Scope = types.StringValue(*destination.AuthMethod.Oauth2ClientCredentials.Config.Scope)
	}
}

func (oauth2ClientCredentialsAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.OAuth2ClientCredentials == nil {
		return nil
	}

	var authenticationType *hookdeck.DestinationAuthMethodOAuth2ClientCredentialsConfigAuthenticationType
	authenticationType = nil
	if !method.OAuth2ClientCredentials.AuthenticationType.IsNull() && !method.OAuth2ClientCredentials.AuthServer.IsUnknown() {
		authenticationTypeValue, err := hookdeck.NewDestinationAuthMethodOAuth2ClientCredentialsConfigAuthenticationTypeFromString(method.OAuth2ClientCredentials.AuthenticationType.ValueString())
		if err != nil {
			panic(err)
		}
		authenticationType = &authenticationTypeValue
	}

	return hookdeck.NewDestinationAuthMethodConfigFromOauth2ClientCredentials(&hookdeck.AuthOAuth2ClientCredentials{
		Config: &hookdeck.DestinationAuthMethodOAuth2ClientCredentialsConfig{
			AuthServer:         method.OAuth2ClientCredentials.AuthServer.ValueString(),
			AuthenticationType: authenticationType,
			ClientId:           method.OAuth2ClientCredentials.ClientID.ValueString(),
			ClientSecret:       method.OAuth2ClientCredentials.ClientSecret.ValueString(),
			Scope:              method.OAuth2ClientCredentials.Scope.ValueStringPointer(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &oauth2ClientCredentialsAuthenticationMethod{})
}
