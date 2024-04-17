package destination

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationResourceModel struct {
	AuthMethod             *destinationAuthMethodConfig `tfsdk:"auth_method"`
	CliPath                types.String                 `tfsdk:"cli_path"`
	CreatedAt              types.String                 `tfsdk:"created_at"`
	Description            types.String                 `tfsdk:"description"`
	DisabledAt             types.String                 `tfsdk:"disabled_at"`
	HTTPMethod             types.String                 `tfsdk:"http_method"`
	ID                     types.String                 `tfsdk:"id"`
	Name                   types.String                 `tfsdk:"name"`
	PathForwardingDisabled types.Bool                   `tfsdk:"path_forwarding_disabled"`
	RateLimit              *rateLimit                   `tfsdk:"rate_limit"`
	TeamID                 types.String                 `tfsdk:"team_id"`
	UpdatedAt              types.String                 `tfsdk:"updated_at"`
	URL                    types.String                 `tfsdk:"url"`
}

type rateLimit struct {
	Limit  types.Int64  `tfsdk:"limit"`
	Period types.String `tfsdk:"period"`
}

type destinationAuthMethodConfig struct {
	APIKey            *apiKey            `tfsdk:"api_key"`
	BasicAuth         *basicAuth         `tfsdk:"basic_auth"`
	BearerToken       *bearerToken       `tfsdk:"bearer_token"`
	CustomSignature   *customSignature   `tfsdk:"custom_signature"`
	HookdeckSignature *hookdeckSignature `tfsdk:"hookdeck_signature"`
}

func (destinationAuthMethodConfig) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key":            types.ObjectType{AttrTypes: apiKey{}.attrTypes()},
		"basic_auth":         types.ObjectType{AttrTypes: basicAuth{}.attrTypes()},
		"bearer_token":       types.ObjectType{AttrTypes: bearerToken{}.attrTypes()},
		"custom_signature":   types.ObjectType{AttrTypes: customSignature{}.attrTypes()},
		"hookdeck_signature": types.ObjectType{AttrTypes: hookdeckSignature{}.attrTypes()},
	}
}

func (destinationAuthMethodConfig) defaultValue() map[string]attr.Value {
	return map[string]attr.Value{
		"api_key":            apiKey{}.defaultValue(),
		"basic_auth":         basicAuth{}.defaultValue(),
		"bearer_token":       bearerToken{}.defaultValue(),
		"custom_signature":   customSignature{}.defaultValue(),
		"hookdeck_signature": hookdeckSignature{}.defaultValue(),
	}
}

type apiKey struct {
	APIKey types.String `tfsdk:"api_key"`
	Key    types.String `tfsdk:"key"`
	To     types.String `tfsdk:"to"`
}

func (apiKey) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"api_key": types.StringType,
		"key":     types.StringType,
		"to":      types.StringType,
	}
}

func (m apiKey) defaultValue() attr.Value {
	return types.ObjectNull(m.attrTypes())
}

type basicAuth struct {
	Password types.String `tfsdk:"password"`
	Username types.String `tfsdk:"username"`
}

func (basicAuth) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"password": types.StringType,
		"username": types.StringType,
	}
}

func (m basicAuth) defaultValue() attr.Value {
	return types.ObjectNull(m.attrTypes())
}

type bearerToken struct {
	Token types.String `tfsdk:"token"`
}

func (bearerToken) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"token": types.StringType,
	}
}

func (m bearerToken) defaultValue() attr.Value {
	return types.ObjectNull(m.attrTypes())
}

type customSignature struct {
	Key           types.String `tfsdk:"key"`
	SigningSecret types.String `tfsdk:"signing_secret"`
}

func (customSignature) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":            types.StringType,
		"signing_secret": types.StringType,
	}
}

func (m customSignature) defaultValue() attr.Value {
	return types.ObjectNull(m.attrTypes())
}

type hookdeckSignature struct {
}

func (hookdeckSignature) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{}
}

func (m hookdeckSignature) defaultValue() attr.Value {
	value, _ := types.ObjectValue(m.attrTypes(), map[string]attr.Value{})
	return value
}
