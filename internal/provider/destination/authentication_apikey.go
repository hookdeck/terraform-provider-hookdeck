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

type apiKeyAuthenticationMethodModel struct {
	APIKey types.String `tfsdk:"api_key"`
	Key    types.String `tfsdk:"key"`
	To     types.String `tfsdk:"to"`
}

type apiKeyAuthenticationMethod struct {
}

func (*apiKeyAuthenticationMethod) name() string {
	return "api_key"
}

func (*apiKeyAuthenticationMethod) schema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: `API key for the API key auth`,
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: `Key for the API key auth`,
			},
			"to": schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf("header", "query")},
				Default:    stringdefault.StaticString("header"),
				MarkdownDescription: `must be one of ["header", "query"]` + "\n" +
					`Whether the API key should be sent as a header or a query parameter`,
			},
		},
		Description: `API Key`,
	}
}

func apiKeyAuthenticationMethodAttrTypesMap() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key": types.StringType,
		"key":     types.StringType,
		"to":      types.StringType,
	}
}

func (apiKeyAuthenticationMethod) attrTypes() attr.Type {
	return types.ObjectType{AttrTypes: apiKeyAuthenticationMethodAttrTypesMap()}
}

func (apiKeyAuthenticationMethod) defaultValue() attr.Value {
	return types.ObjectNull(apiKeyAuthenticationMethodAttrTypesMap())
}

func (apiKeyAuthenticationMethod) refresh(m *destinationResourceModel, destination *hookdeck.Destination) {
	if destination.AuthMethod.ApiKey == nil {
		return
	}

	m.AuthMethod.APIKey = &apiKeyAuthenticationMethodModel{}
	m.AuthMethod.APIKey.APIKey = types.StringValue(destination.AuthMethod.ApiKey.Config.ApiKey)
	m.AuthMethod.APIKey.Key = types.StringValue(destination.AuthMethod.ApiKey.Config.Key)
	m.AuthMethod.APIKey.To = types.StringValue(string(*destination.AuthMethod.ApiKey.Config.To.Ptr()))
}

func (apiKeyAuthenticationMethod) toPayload(method *destinationAuthMethodConfig) *hookdeck.DestinationAuthMethodConfig {
	if method.APIKey == nil {
		return nil
	}

	to, _ := hookdeck.NewDestinationAuthMethodApiKeyConfigToFromString(method.APIKey.To.ValueString())
	return hookdeck.NewDestinationAuthMethodConfigFromApiKey(&hookdeck.AuthApiKey{
		Config: &hookdeck.DestinationAuthMethodApiKeyConfig{
			Key:    method.APIKey.Key.ValueString(),
			ApiKey: method.APIKey.APIKey.ValueString(),
			To:     &to,
		},
	})
}

func init() {
	authenticationMethods = append(authenticationMethods, &apiKeyAuthenticationMethod{})
}
