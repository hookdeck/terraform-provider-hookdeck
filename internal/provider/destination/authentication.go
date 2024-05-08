package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type destinationAuthMethodConfig struct {
	APIKey                  *apiKeyAuthenticationMethodModel                  `tfsdk:"api_key"`
	AWSSignature            *awsSignatureAuthenticationMethodModel            `tfsdk:"aws_signature"`
	BasicAuth               *basicAuthAuthenticationMethodModel               `tfsdk:"basic_auth"`
	BearerToken             *bearerTokenAuthenticationMethodModel             `tfsdk:"bearer_token"`
	CustomSignature         *customSignatureAuthenticationMethodModel         `tfsdk:"custom_signature"`
	HookdeckSignature       *hookdeckSignatureAuthenticationMethodModel       `tfsdk:"hookdeck_signature"`
	JSON                    jsonAuthenticationMethodModel                     `tfsdk:"json"`
	OAuth2AuthorizationCode *oauth2AuthorizationCodeAuthenticationMethodModel `tfsdk:"oauth2_authorization_code"`
	OAuth2ClientCredentials *oauth2ClientCredentialsAuthenticationMethodModel `tfsdk:"oauth2_client_credentials"`
}

type authenticationMethod interface {
	name() string
	schema() schema.Attribute
	attrTypes() attr.Type
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
		attrTypes[method.name()] = method.attrTypes()
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
