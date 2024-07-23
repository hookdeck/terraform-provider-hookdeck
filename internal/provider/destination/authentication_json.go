package destination

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type jsonAuthenticationMethodModel = types.String

type jsonAuthenticationMethod struct {
}

func (*jsonAuthenticationMethod) name() string {
	return "json"
}

func (*jsonAuthenticationMethod) schema() schema.Attribute {
	return schema.StringAttribute{
		Optional:    true,
		Sensitive:   true,
		Description: `Stringified JSON value for destination payload, used when Terraform provider hasn't supported the destination method on Hookdeck yet`,
	}
}

func (jsonAuthenticationMethod) attrTypes() attr.Type {
	return types.StringType
}

func (jsonAuthenticationMethod) defaultValue() attr.Value {
	return types.StringNull()
}

func (jsonAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
}

func (jsonAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.JSON.IsNull() || method.JSON.IsUnknown() {
		return nil
	}

	var authenticationMethodConfig *hookdeck.DestinationAuthMethodConfig
	if err := json.Unmarshal([]byte(method.JSON.ValueString()), &authenticationMethodConfig); err != nil {
		// TODO: improve error handling?
		log.Fatal("Error unmarshalling JSON source verification payload")
	}
	return authenticationMethodConfig
}

func init() {
	authenticationMethods = append(authenticationMethods, &jsonAuthenticationMethod{})
}
