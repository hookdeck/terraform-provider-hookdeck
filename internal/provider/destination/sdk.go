package destination

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *destinationResourceModel) Refresh(destination *hookdeck.Destination) {
	if destination.ArchivedAt != nil {
		m.ArchivedAt = types.StringValue(destination.ArchivedAt.Format(time.RFC3339))
	} else {
		m.ArchivedAt = types.StringNull()
	}
	// if destination.AuthMethod == nil {
	// 	m.AuthMethod = nil
	// } else {
	// 	m.AuthMethod = &destinationAuthMethodConfig{}
	// 	if destination.AuthMethod.ApiKey != nil {
	// 		m.AuthMethod.APIKey = &APIKey{}
	// 		if m.AuthMethod.APIKey.Config == nil {
	// 			m.AuthMethod.APIKey.Config = &destinationAuthMethodAPIKeyConfig{}
	// 		}
	// 		if destination.AuthMethod.ApiKey.Configs == nil {
	// 			m.AuthMethod.APIKey.Config = nil
	// 		} else {
	// 			m.AuthMethod.APIKey.Config = &destinationAuthMethodAPIKeyConfig{}
	// 			m.AuthMethod.APIKey.Config.APIKey = types.StringValue(destination.AuthMethod.ApiKey.Configs.ApiKey)
	// 			m.AuthMethod.APIKey.Config.Key = types.StringValue(destination.AuthMethod.ApiKey.Configs.Key)
	// 			if destination.AuthMethod.ApiKey.Configs.To != nil {
	// 				m.AuthMethod.APIKey.Config.To = types.StringValue(string(*destination.AuthMethod.ApiKey.Configs.To))
	// 			} else {
	// 				m.AuthMethod.APIKey.Config.To = types.StringNull()
	// 			}
	// 		}
	// 		m.AuthMethod.APIKey.Type = types.StringValue(string(destination.AuthMethod.ApiKey.Type()))
	// 	}
	// 	if destination.AuthMethod.BasicAuth != nil {
	// 		m.AuthMethod.BasicAuth = &basicAuth{}
	// 		if m.AuthMethod.BasicAuth.Config == nil {
	// 			m.AuthMethod.BasicAuth.Config = &destinationAuthMethodBasicAuthConfig{}
	// 		}
	// 		if destination.AuthMethod.BasicAuth.Configs == nil {
	// 			m.AuthMethod.BasicAuth.Config = nil
	// 		} else {
	// 			m.AuthMethod.BasicAuth.Config = &destinationAuthMethodBasicAuthConfig{}
	// 			m.AuthMethod.BasicAuth.Config.Password = types.StringValue(destination.AuthMethod.BasicAuth.Configs.Password)
	// 			m.AuthMethod.BasicAuth.Config.Username = types.StringValue(destination.AuthMethod.BasicAuth.Configs.Name)
	// 		}
	// 		m.AuthMethod.BasicAuth.Type = types.StringValue(string(destination.AuthMethod.BasicAuth.Type()))
	// 	}
	// 	if destination.AuthMethod.BearerToken != nil {
	// 		m.AuthMethod.BearerToken = &bearerToken{}
	// 		if m.AuthMethod.BearerToken.Config == nil {
	// 			m.AuthMethod.BearerToken.Config = &destinationAuthMethodBearerTokenConfig{}
	// 		}
	// 		if destination.AuthMethod.BearerToken.Config == nil {
	// 			m.AuthMethod.BearerToken.Config = nil
	// 		} else {
	// 			m.AuthMethod.BearerToken.Config = &destinationAuthMethodBearerTokenConfig{}
	// 			m.AuthMethod.BearerToken.Config.Token = types.StringValue(destination.AuthMethod.BearerToken.Config.Token)
	// 		}
	// 		m.AuthMethod.BearerToken.Type = types.StringValue(string(destination.AuthMethod.BearerToken.Type()))
	// 	}
	// 	if destination.AuthMethod.CustomSignature != nil {
	// 		m.AuthMethod.CustomSignature = &customSignature{}
	// 		m.AuthMethod.CustomSignature.Config.Key = types.StringValue(destination.AuthMethod.CustomSignature.Config.Key)
	// 		if destination.AuthMethod.CustomSignature.Config.SigningSecret != nil {
	// 			m.AuthMethod.CustomSignature.Config.SigningSecret = types.StringValue(*destination.AuthMethod.CustomSignature.Config.SigningSecret)
	// 		} else {
	// 			m.AuthMethod.CustomSignature.Config.SigningSecret = types.StringNull()
	// 		}
	// 		m.AuthMethod.CustomSignature.Type = types.StringValue(string(destination.AuthMethod.CustomSignature.Type()))
	// 	}
	// 	if destination.AuthMethod.HookdeckSignature != nil {
	// 		m.AuthMethod.HookdeckSignature = &HookdeckSignature{}
	// 		if m.AuthMethod.HookdeckSignature.Config == nil {
	// 			m.AuthMethod.HookdeckSignature.Config = &destinationAuthMethodHookdeckSignatureConfig{}
	// 		}
	// 		if destination.AuthMethod.HookdeckSignature.Config == nil {
	// 			m.AuthMethod.HookdeckSignature.Config = nil
	// 		} else {
	// 			m.AuthMethod.HookdeckSignature.Config = &destinationAuthMethodHookdeckSignatureConfig{}
	// 		}
	// 		m.AuthMethod.HookdeckSignature.Type = types.StringValue(string(destination.AuthMethod.HookdeckSignature.Type()))
	// 	}
	// }
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
		PathForwardingDisabled: hookdeck.OptionalOrNull(pathForwardingDisabled),
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
	return nil
	// var authMethod *hookdeck.DestinationAuthMethodConfig
	// if m.AuthMethod != nil {
	// 	var hookdeckSignature *hookdeck.HookdeckSignature
	// 	if m.AuthMethod.HookdeckSignature != nil {
	// 		var config *hookdeck.DestinationAuthMethodSignatureConfig
	// 		if m.AuthMethod.HookdeckSignature.Config != nil {
	// 			config = &hookdeck.DestinationAuthMethodSignatureConfig{}
	// 		}
	// 		hookdeckSignature = &hookdeck.HookdeckSignature{
	// 			Config: config,
	// 		}
	// 	}
	// 	if hookdeckSignature != nil {
	// 		authMethod = &hookdeck.DestinationAuthMethodConfig{
	// 			HookdeckSignature: hookdeckSignature,
	// 		}
	// 	}
	// 	var basicAuth *hookdeck.BasicAuth
	// 	if m.AuthMethod.BasicAuth != nil {
	// 		var config *hookdeck.BasicAuthConfigs
	// 		if m.AuthMethod.BasicAuth.Config != nil {
	// 			config = &hookdeck.BasicAuthConfigs{
	// 				Password: m.AuthMethod.BasicAuth.Config.Password.ValueString(),
	// 				Name:     m.AuthMethod.BasicAuth.Config.Username.ValueString(),
	// 			}
	// 		}
	// 		basicAuth = &hookdeck.BasicAuth{
	// 			Configs: config,
	// 		}
	// 	}
	// 	if basicAuth != nil {
	// 		authMethod = &hookdeck.DestinationAuthMethodConfig{
	// 			BasicAuth: basicAuth,
	// 		}
	// 	}
	// 	// var apiKey *hookdeck.ApiKey
	// 	// if m.AuthMethod.APIKey != nil {
	// 	// 	var config2 *hookdeck.DestinationAuthMethodAPIKeyConfig
	// 	// 	if m.AuthMethod.APIKey.Config != nil {
	// 	// 		apiKey1 := m.AuthMethod.APIKey.Config.APIKey.ValueString()
	// 	// 		key := m.AuthMethod.APIKey.Config.Key.ValueString()
	// 	// 		to := new(hookdeck.DestinationAuthMethodAPIKeyConfigTo)
	// 	// 		if !m.AuthMethod.APIKey.Config.To.IsUnknown() && !m.AuthMethod.APIKey.Config.To.IsNull() {
	// 	// 			*to = hookdeck.DestinationAuthMethodAPIKeyConfigTo(m.AuthMethod.APIKey.Config.To.ValueString())
	// 	// 		} else {
	// 	// 			to = nil
	// 	// 		}
	// 	// 		config2 = &hookdeck.DestinationAuthMethodAPIKeyConfig{
	// 	// 			APIKey: apiKey1,
	// 	// 			Key:    key,
	// 	// 			To:     to,
	// 	// 		}
	// 	// 	}
	// 	// 	typeVar2 := hookdeck.APIKeyType(m.AuthMethod.APIKey.Type.ValueString())
	// 	// 	apiKey = &hookdeck.APIKey{
	// 	// 		Config: config2,
	// 	// 		Type:   typeVar2,
	// 	// 	}
	// 	// }
	// 	// if apiKey != nil {
	// 	// 	authMethod = &hookdeck.DestinationAuthMethodConfig{
	// 	// 		APIKey: apiKey,
	// 	// 	}
	// 	// }
	// 	var bearerToken *hookdeck.BearerToken
	// 	if m.AuthMethod.BearerToken != nil {
	// 		var config *hookdeck.DestinationAuthMethodBearerTokenConfig
	// 		if m.AuthMethod.BearerToken.Config != nil {
	// 			token := m.AuthMethod.BearerToken.Config.Token.ValueString()
	// 			config = &hookdeck.DestinationAuthMethodBearerTokenConfig{
	// 				Token: token,
	// 			}
	// 		}
	// 		bearerToken = &hookdeck.BearerToken{
	// 			Config: config,
	// 		}
	// 	}
	// 	if bearerToken != nil {
	// 		authMethod = &hookdeck.DestinationAuthMethodConfig{
	// 			BearerToken: bearerToken,
	// 		}
	// 	}
	// 	var customSignature *hookdeck.CustomSignature
	// 	if m.AuthMethod.CustomSignature != nil {
	// 		key1 := m.AuthMethod.CustomSignature.Config.Key.ValueString()
	// 		signingSecret := new(string)
	// 		if !m.AuthMethod.CustomSignature.Config.SigningSecret.IsUnknown() && !m.AuthMethod.CustomSignature.Config.SigningSecret.IsNull() {
	// 			*signingSecret = m.AuthMethod.CustomSignature.Config.SigningSecret.ValueString()
	// 		} else {
	// 			signingSecret = nil
	// 		}
	// 		config := hookdeck.DestinationAuthMethodCustomSignatureConfig{
	// 			Key:           key1,
	// 			SigningSecret: signingSecret,
	// 		}
	// 		customSignature = &hookdeck.CustomSignature{
	// 			Config: &config,
	// 		}
	// 	}
	// 	if customSignature != nil {
	// 		authMethod = &hookdeck.DestinationAuthMethodConfig{
	// 			CustomSignature: customSignature,
	// 		}
	// 	}
	// }

	// return authMethod
}
