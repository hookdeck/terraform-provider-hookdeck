package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type basicAuthAuthenticationMethodModel struct {
	Password types.String `tfsdk:"password"`
	Username types.String `tfsdk:"username"`
}

type basicAuthAuthenticationMethod struct {
}

func (*basicAuthAuthenticationMethod) name() string {
	return "basic_auth"
}

func (*basicAuthAuthenticationMethod) schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `Password for basic auth`,
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: `Username for basic auth`,
			},
		},
		Description: `Basic Auth`,
	}
}

func (basicAuthAuthenticationMethod) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"password": types.StringType,
		"username": types.StringType,
	}
}

func (basicAuthAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(basicAuthAuthenticationMethod{}.attrTypes())
}

func (basicAuthAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.BasicAuth == nil {
		return
	}

	m.AuthMethod.BasicAuth = &basicAuthAuthenticationMethodModel{}
	m.AuthMethod.BasicAuth.Password = types.StringValue(destination.AuthMethod.BasicAuth.Config.Password)
	m.AuthMethod.BasicAuth.Username = types.StringValue(destination.AuthMethod.BasicAuth.Config.Username)
}

func (basicAuthAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.BasicAuth == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromBasicAuth(&hookdeck.AuthBasicAuth{
		Config: &hookdeck.DestinationAuthMethodBasicAuthConfig{
			Password: method.BasicAuth.Password.ValueString(),
			Username: method.BasicAuth.Username.ValueString(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &basicAuthAuthenticationMethod{})
}
