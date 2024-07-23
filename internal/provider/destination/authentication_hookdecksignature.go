package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type hookdeckSignatureAuthenticationMethodModel struct {
}

type hookdeckSignatureAuthenticationMethod struct {
}

func (*hookdeckSignatureAuthenticationMethod) name() string {
	return "hookdeck_signature"
}

func (*hookdeckSignatureAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Attributes:  map[string]schema.Attribute{},
		Description: `Hookdeck Signature`,
	}
}

func hookdeckSignatureAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{}
}

func (hookdeckSignatureAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: hookdeckSignatureAuthenticationMethodAttrTypesMap()}
}

func (hookdeckSignatureAuthenticationMethod) defaultValue() attr.Value {
	value, _ := types.ObjectValue(hookdeckSignatureAuthenticationMethodAttrTypesMap(), map[string]attr.Value{})
	return value
}

func (hookdeckSignatureAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.HookdeckSignature == nil {
		return
	}

	m.AuthMethod.HookdeckSignature = &hookdeckSignatureAuthenticationMethodModel{}
}

func (hookdeckSignatureAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.HookdeckSignature == nil {
		return nil
	}

	return hookdeck.NewDestinationAuthMethodConfigFromHookdeckSignature(&hookdeck.AuthHookdeckSignature{
		Config: &hookdeck.DestinationAuthMethodSignatureConfig{},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &hookdeckSignatureAuthenticationMethod{})
}
