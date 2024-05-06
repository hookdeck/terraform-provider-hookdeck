package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type customSignatureAuthenticationMethodModel struct {
	Key           types.String `tfsdk:"key"`
	SigningSecret types.String `tfsdk:"signing_secret"`
}

type customSignatureAuthenticationMethod struct {
}

func (*customSignatureAuthenticationMethod) name() string {
	return "custom_signature"
}

func (*customSignatureAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:    true,
				Description: `Key for the custom signature auth`,
			},
			"signing_secret": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Sensitive:   true,
				Description: `Signing secret for the custom signature auth. If left empty a secret will be generated for you.`,
			},
		},
		Description: `Custom Signature`,
	}
}

func customSignatureAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"key":            types.StringType,
		"signing_secret": types.StringType,
	}
}

func (customSignatureAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: customSignatureAuthenticationMethodAttrTypesMap()}
}

func (customSignatureAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(customSignatureAuthenticationMethodAttrTypesMap())
}

func (customSignatureAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.CustomSignature == nil {
		return
	}

	m.AuthMethod.CustomSignature = &customSignatureAuthenticationMethodModel{}
	m.AuthMethod.CustomSignature.Key = types.StringValue(destination.AuthMethod.CustomSignature.Config.Key)
	m.AuthMethod.CustomSignature.SigningSecret = types.StringValue(*destination.AuthMethod.CustomSignature.Config.SigningSecret)
}

func (customSignatureAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.CustomSignature == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromCustomSignature(&hookdeck.AuthCustomSignature{
		Config: &hookdeck.DestinationAuthMethodCustomSignatureConfig{
			Key:           method.CustomSignature.Key.ValueString(),
			SigningSecret: method.CustomSignature.SigningSecret.ValueStringPointer(),
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &customSignatureAuthenticationMethod{})
}
