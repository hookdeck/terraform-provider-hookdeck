package destination

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type jsonAuthenticationMethodModel struct {
	JSON types.String `tfsdk:"json"`
}

type jsonAuthenticationMethod struct {
}

func (*jsonAuthenticationMethod) name() string {
	return "json"
}

func (*jsonAuthenticationMethod) schema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"json": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `Stringified JSON value for destination payload`,
			},
		},
		Description: `Used when Terraform provider hasn't supported the destination method on Hookdeck yet`,
	}
}

func (jsonAuthenticationMethod) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"json": types.StringType,
	}
}

func (jsonAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(jsonAuthenticationMethod{}.attrTypes())
}

func (jsonAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
}

func (jsonAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.JSON == nil {
		return nil
	}

	var authenticationMethodConfig *hookdeck.DestinationAuthMethodConfig
	if err := json.Unmarshal([]byte(method.JSON.JSON.ValueString()), &authenticationMethodConfig); err != nil {
		// TODO: improve error handling?
		log.Fatal("Error unmarshalling JSON source verification payload")
	}
	return authenticationMethodConfig
}

func init() {
	authenticationMethods = append(authenticationMethods, &jsonAuthenticationMethod{})
}
