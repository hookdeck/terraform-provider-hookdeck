package destination

import "github.com/hashicorp/terraform-plugin-framework/types"

type destinationResourceModel struct {
	ArchivedAt types.String `tfsdk:"archived_at"`
	// AuthMethod             *destinationAuthMethodConfig `tfsdk:"auth_method"`
	CliPath                types.String `tfsdk:"cli_path"`
	CreatedAt              types.String `tfsdk:"created_at"`
	Description            types.String `tfsdk:"description"`
	HTTPMethod             types.String `tfsdk:"http_method"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	PathForwardingDisabled types.Bool   `tfsdk:"path_forwarding_disabled"`
	RateLimit              types.Int64  `tfsdk:"rate_limit"`
	RateLimitPeriod        types.String `tfsdk:"rate_limit_period"`
	TeamID                 types.String `tfsdk:"team_id"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	URL                    types.String `tfsdk:"url"`
}

// type destinationAuthMethodConfig struct {
// 	APIKey            *APIKey            `tfsdk:"api_key"`
// 	BasicAuth         *basicAuth         `tfsdk:"basic_auth"`
// 	BearerToken       *bearerToken       `tfsdk:"bearer_token"`
// 	CustomSignature   *customSignature   `tfsdk:"custom_signature"`
// 	HookdeckSignature *HookdeckSignature `tfsdk:"hookdeck_signature"`
// }

// type APIKey struct {
// 	Config *destinationAuthMethodAPIKeyConfig `tfsdk:"config"`
// 	Type   types.String                       `tfsdk:"type"`
// }

// type destinationAuthMethodAPIKeyConfig struct {
// 	APIKey types.String `tfsdk:"api_key"`
// 	Key    types.String `tfsdk:"key"`
// 	To     types.String `tfsdk:"to"`
// }

// type basicAuth struct {
// 	Config *destinationAuthMethodBasicAuthConfig `tfsdk:"config"`
// 	Type   types.String                          `tfsdk:"type"`
// }

// type destinationAuthMethodBasicAuthConfig struct {
// 	Password types.String `tfsdk:"password"`
// 	Username types.String `tfsdk:"username"`
// }

// type bearerToken struct {
// 	Config *destinationAuthMethodBearerTokenConfig `tfsdk:"config"`
// 	Type   types.String                            `tfsdk:"type"`
// }

// type destinationAuthMethodBearerTokenConfig struct {
// 	Token types.String `tfsdk:"token"`
// }

// type customSignature struct {
// 	Config destinationAuthMethodCustomSignatureConfig `tfsdk:"config"`
// 	Type   types.String                               `tfsdk:"type"`
// }

// type destinationAuthMethodCustomSignatureConfig struct {
// 	Key           types.String `tfsdk:"key"`
// 	SigningSecret types.String `tfsdk:"signing_secret"`
// }

// type HookdeckSignature struct {
// 	Config *destinationAuthMethodHookdeckSignatureConfig `tfsdk:"config"`
// 	Type   types.String                                  `tfsdk:"type"`
// }

// type destinationAuthMethodHookdeckSignatureConfig struct {
// }
