package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type authenticationMethod interface {
	name() string
	schema() schema.SingleNestedAttribute
	attrTypes() map[string]attr.Type
	defaultValue() attr.Value

	// SDK methods
	refresh(m *destinationResourceModel, destination *hookdeck.Destination)
	toPayload(authMethod *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig
}

var authenticationMethods []authenticationMethod

func getAuthenticationMethodSchemaAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{}

	for _, method := range authenticationMethods {
		attributes[method.name()] = method.schema()
	}

	return attributes
}

func getAuthenticationMethodSchemaAttrTypes() map[string]attr.Type {
	attrTypes := map[string]attr.Type{}

	for _, method := range authenticationMethods {
		attrTypes[method.name()] = types.ObjectType{AttrTypes: method.attrTypes()}
	}

	return attrTypes
}

func getAuthenticationMethodSchemaDefaultValue() map[string]attr.Value {
	defaultValues := map[string]attr.Value{}

	for _, method := range authenticationMethods {
		defaultValues[method.name()] = method.defaultValue()
	}

	return defaultValues
}
