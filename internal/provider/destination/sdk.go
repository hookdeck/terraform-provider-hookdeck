package destination

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckcore "github.com/hookdeck/hookdeck-go-sdk/core"
)

func (m *destinationResourceModel) Refresh(destination *hookdeck.Destination) {
	if destination.ArchivedAt != nil {
		m.ArchivedAt = types.StringValue(destination.ArchivedAt.Format(time.RFC3339))
	} else {
		m.ArchivedAt = types.StringNull()
	}
	if destination.AuthMethod != nil {
		m.refreshAuthMethod(destination)
	}
	if destination.CliPath != nil {
		m.CliPath = types.StringValue(*destination.CliPath)
	} else {
		m.CliPath = types.StringNull()
	}
	m.CreatedAt = types.StringValue(destination.CreatedAt.Format(time.RFC3339))
	if destination.HttpMethod != nil {
		m.HTTPMethod = types.StringValue(string(*destination.HttpMethod))
	} else {
		m.HTTPMethod = types.StringNull()
	}
	m.ID = types.StringValue(destination.Id)
	m.Name = types.StringValue(destination.Name)
	if destination.PathForwardingDisabled != nil {
		m.PathForwardingDisabled = types.BoolValue(*destination.PathForwardingDisabled)
	} else {
		m.PathForwardingDisabled = types.BoolNull()
	}
	if destination.RateLimit != nil {
		m.RateLimit = types.Int64Value(int64(*destination.RateLimit))
	} else {
		m.RateLimit = types.Int64Null()
	}
	if destination.RateLimitPeriod != nil {
		m.RateLimitPeriod = types.StringValue(string(*destination.RateLimitPeriod))
	} else {
		m.RateLimitPeriod = types.StringNull()
	}
	m.TeamID = types.StringValue(destination.TeamId)
	m.UpdatedAt = types.StringValue(destination.UpdatedAt.Format(time.RFC3339))
	if destination.Url != nil {
		m.URL = types.StringValue(*destination.Url)
	} else {
		m.URL = types.StringNull()
	}
}

func (m *destinationResourceModel) ToCreatePayload() *hookdeck.DestinationCreateRequest {
	authMethod := m.getAuthMethod()
	cliPath := new(string)
	if !m.CliPath.IsUnknown() && !m.CliPath.IsNull() {
		*cliPath = m.CliPath.ValueString()
	} else {
		cliPath = nil
	}
	httpMethod := new(hookdeck.DestinationHttpMethod)
	if !m.HTTPMethod.IsUnknown() && !m.HTTPMethod.IsNull() {
		*httpMethod = hookdeck.DestinationHttpMethod(m.HTTPMethod.ValueString())
	} else {
		httpMethod = nil
	}
	var pathForwardingDisabled *hookdeckcore.Optional[bool] = nil
	if !m.PathForwardingDisabled.IsUnknown() && !m.PathForwardingDisabled.IsNull() {
		pathForwardingDisabled = hookdeck.Optional(m.PathForwardingDisabled.ValueBool())
	} else {
		pathForwardingDisabled = nil
	}
	rateLimit := new(int)
	if !m.RateLimit.IsUnknown() && !m.RateLimit.IsNull() {
		*rateLimit = int(m.RateLimit.ValueInt64())
	} else {
		rateLimit = nil
	}
	rateLimitPeriod := new(hookdeck.DestinationCreateRequestRateLimitPeriod)
	if rateLimit != nil {
		if !m.RateLimitPeriod.IsUnknown() && !m.RateLimitPeriod.IsNull() {
			*rateLimitPeriod = hookdeck.DestinationCreateRequestRateLimitPeriod(m.RateLimitPeriod.ValueString())
		} else {
			rateLimitPeriod = nil
		}
	} else {
		rateLimitPeriod = nil
	}
	url := new(string)
	if !m.URL.IsUnknown() && !m.URL.IsNull() {
		*url = m.URL.ValueString()
	} else {
		url = nil
	}
	return &hookdeck.DestinationCreateRequest{
		AuthMethod:             hookdeck.OptionalOrNull(authMethod),
		CliPath:                hookdeck.OptionalOrNull(cliPath),
		Description:            hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		HttpMethod:             hookdeck.OptionalOrNull(httpMethod),
		Name:                   m.Name.ValueString(),
		PathForwardingDisabled: pathForwardingDisabled,
		RateLimit:              hookdeck.OptionalOrNull(rateLimit),
		RateLimitPeriod:        hookdeck.OptionalOrNull(rateLimitPeriod),
		Url:                    hookdeck.OptionalOrNull(url),
	}
}

func (m *destinationResourceModel) ToUpdatePayload() *hookdeck.DestinationUpdateRequest {
	authMethod := m.getAuthMethod()
	cliPath := new(string)
	if !m.CliPath.IsUnknown() && !m.CliPath.IsNull() {
		*cliPath = m.CliPath.ValueString()
	} else {
		cliPath = nil
	}
	httpMethod := new(hookdeck.DestinationHttpMethod)
	if !m.HTTPMethod.IsUnknown() && !m.HTTPMethod.IsNull() {
		*httpMethod = hookdeck.DestinationHttpMethod(m.HTTPMethod.ValueString())
	} else {
		httpMethod = nil
	}
	pathForwardingDisabled := new(bool)
	if !m.PathForwardingDisabled.IsUnknown() && !m.PathForwardingDisabled.IsNull() {
		*pathForwardingDisabled = m.PathForwardingDisabled.ValueBool()
	} else {
		pathForwardingDisabled = nil
	}
	rateLimit := new(int)
	if !m.RateLimit.IsUnknown() && !m.RateLimit.IsNull() {
		*rateLimit = int(m.RateLimit.ValueInt64())
	} else {
		rateLimit = nil
	}
	rateLimitPeriod := new(hookdeck.DestinationUpdateRequestRateLimitPeriod)
	if rateLimit != nil {
		if !m.RateLimitPeriod.IsUnknown() && !m.RateLimitPeriod.IsNull() {
			*rateLimitPeriod = hookdeck.DestinationUpdateRequestRateLimitPeriod(m.RateLimitPeriod.ValueString())
		} else {
			rateLimitPeriod = nil
		}
	} else {
		rateLimitPeriod = nil
	}
	url := new(string)
	if !m.URL.IsUnknown() && !m.URL.IsNull() {
		*url = m.URL.ValueString()
	} else {
		url = nil
	}
	return &hookdeck.DestinationUpdateRequest{
		AuthMethod:             hookdeck.OptionalOrNull(authMethod),
		CliPath:                hookdeck.OptionalOrNull(cliPath),
		Description:            hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		HttpMethod:             hookdeck.OptionalOrNull(httpMethod),
		Name:                   hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
		PathForwardingDisabled: hookdeck.OptionalOrNull(pathForwardingDisabled),
		RateLimit:              hookdeck.OptionalOrNull(rateLimit),
		RateLimitPeriod:        hookdeck.OptionalOrNull(rateLimitPeriod),
		Url:                    hookdeck.OptionalOrNull(url),
	}
}

func (m *destinationResourceModel) getAuthMethod() *hookdeck.DestinationAuthMethodConfig {
	if m.AuthMethod == nil {
		return nil
	}

	if m.AuthMethod.APIKey != nil {
		to, _ := hookdeck.NewDestinationAuthMethodApiKeyConfigToFromString(m.AuthMethod.APIKey.To.ValueString())
		return hookdeck.NewDestinationAuthMethodConfigFromApiKey(&hookdeck.AuthApiKey{
			Config: &hookdeck.DestinationAuthMethodApiKeyConfig{
				Key:    m.AuthMethod.APIKey.Key.ValueString(),
				ApiKey: m.AuthMethod.APIKey.APIKey.ValueString(),
				To:     &to,
			},
		})
	}

	if m.AuthMethod.BasicAuth != nil {
		return hookdeck.NewDestinationAuthMethodConfigFromBasicAuth(&hookdeck.AuthBasicAuth{
			Config: &hookdeck.DestinationAuthMethodBasicAuthConfig{
				Password: m.AuthMethod.BasicAuth.Password.ValueString(),
				Username: m.AuthMethod.BasicAuth.Username.ValueString(),
			},
		})
	}

	if m.AuthMethod.BearerToken != nil {
		return hookdeck.NewDestinationAuthMethodConfigFromBearerToken(&hookdeck.AuthBearerToken{
			Config: &hookdeck.DestinationAuthMethodBearerTokenConfig{
				Token: m.AuthMethod.BearerToken.Token.ValueString(),
			},
		})
	}

	if m.AuthMethod.CustomSignature != nil {
		return hookdeck.NewDestinationAuthMethodConfigFromCustomSignature(&hookdeck.AuthCustomSignature{
			Config: &hookdeck.DestinationAuthMethodCustomSignatureConfig{
				Key:           m.AuthMethod.CustomSignature.Key.ValueString(),
				SigningSecret: m.AuthMethod.CustomSignature.SigningSecret.ValueStringPointer(),
			},
		})
	}

	if m.AuthMethod.HookdeckSignature != nil {
		return hookdeck.NewDestinationAuthMethodConfigFromHookdeckSignature(&hookdeck.AuthHookdeckSignature{
			Config: &hookdeck.DestinationAuthMethodSignatureConfig{},
		})
	}

	return nil
}

func (m *destinationResourceModel) refreshAuthMethod(destination *hookdeck.Destination) {
	if destination.AuthMethod == nil {
		return
	}

	m.AuthMethod = &destinationAuthMethodConfig{}

	if destination.AuthMethod.ApiKey != nil {
		m.AuthMethod.APIKey = &apiKey{}
		m.AuthMethod.APIKey.APIKey = types.StringValue(destination.AuthMethod.ApiKey.Config.ApiKey)
		m.AuthMethod.APIKey.Key = types.StringValue(destination.AuthMethod.ApiKey.Config.Key)
		m.AuthMethod.APIKey.To = types.StringValue(string(*destination.AuthMethod.ApiKey.Config.To.Ptr()))
	}

	if destination.AuthMethod.BasicAuth != nil {
		m.AuthMethod.BasicAuth = &basicAuth{}
		m.AuthMethod.BasicAuth.Password = types.StringValue(destination.AuthMethod.BasicAuth.Config.Password)
		m.AuthMethod.BasicAuth.Username = types.StringValue(destination.AuthMethod.BasicAuth.Config.Username)
	}

	if destination.AuthMethod.BearerToken != nil {
		m.AuthMethod.BearerToken = &bearerToken{}
		m.AuthMethod.BearerToken.Token = types.StringValue(destination.AuthMethod.BearerToken.Config.Token)
	}

	if destination.AuthMethod.CustomSignature != nil {
		m.AuthMethod.CustomSignature = &customSignature{}
		m.AuthMethod.CustomSignature.Key = types.StringValue(destination.AuthMethod.CustomSignature.Config.Key)
		m.AuthMethod.CustomSignature.SigningSecret = types.StringValue(string(*destination.AuthMethod.CustomSignature.Config.SigningSecret))
	}

	if destination.AuthMethod.HookdeckSignature != nil {
		m.AuthMethod.HookdeckSignature = &hookdeckSignature{}
	}
}
