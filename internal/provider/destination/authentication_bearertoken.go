package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type bearerTokenAuthenticationMethodModel struct {
	Token types.String `tfsdk:"token"`
}

type bearerTokenAuthenticationMethod struct {
}

func (*bearerTokenAuthenticationMethod) name() string {
	return "bearer_token"
}

func (*bearerTokenAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `Token for the bearer token auth`,
			},
		},
		Description: `Bearer Token`,
	}
}

func bearerTokenAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"token": types.StringType,
	}
}

func (bearerTokenAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: bearerTokenAuthenticationMethodAttrTypesMap()}
}

func (bearerTokenAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(bearerTokenAuthenticationMethodAttrTypesMap())
}

func (bearerTokenAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.BearerToken == nil {
		return
	}

	m.AuthMethod.BearerToken = &bearerTokenAuthenticationMethodModel{}
	m.AuthMethod.BearerToken.Token = types.StringValue(destination.AuthMethod.BearerToken.Config.Token)
}

func (bearerTokenAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.BearerToken == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromBearerToken(&hookdeck.AuthBearerToken{
		Config: &hookdeck.DestinationAuthMethodBearerTokenConfig{
			Token: method.BearerToken.Token.ValueString(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &bearerTokenAuthenticationMethod{})
}
