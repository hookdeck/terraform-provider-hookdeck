package destination

import (
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
	APIKey            *apiKeyAuthenticationMethodModel            `tfsdk:"api_key"`
	BasicAuth         *basicAuthAuthenticationMethodModel         `tfsdk:"basic_auth"`
	BearerToken       *bearerTokenAuthenticationMethodModel       `tfsdk:"bearer_token"`
	CustomSignature   *customSignatureAuthenticationMethodModel   `tfsdk:"custom_signature"`
	HookdeckSignature *hookdeckSignatureAuthenticationMethodModel `tfsdk:"hookdeck_signature"`
}
