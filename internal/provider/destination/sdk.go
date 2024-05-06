package destination

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckcore "github.com/hookdeck/hookdeck-go-sdk/core"
)

func (m *destinationResourceModel) Refresh(destination *hookdeck.Destination) {
	if destination.AuthMethod != nil {
		m.refreshAuthMethod(destination)
	}
	if destination.CliPath != nil {
		m.CliPath = types.StringValue(*destination.CliPath)
	} else {
		m.CliPath = types.StringNull()
	}
	m.CreatedAt = types.StringValue(destination.CreatedAt.Format(time.RFC3339))
	if destination.DisabledAt != nil {
		m.DisabledAt = types.StringValue(destination.DisabledAt.Format(time.RFC3339))
	} else {
		m.DisabledAt = types.StringNull()
	}
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
		m.RateLimit = &rateLimit{}
		m.RateLimit.Limit = types.Int64Value(int64(*destination.RateLimit))
		m.RateLimit.Period = types.StringValue(string(*destination.RateLimitPeriod))
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
	var pathForwardingDisabled *hookdeckcore.Optional[bool]
	if !m.PathForwardingDisabled.IsUnknown() && !m.PathForwardingDisabled.IsNull() {
		pathForwardingDisabled = hookdeck.Optional(m.PathForwardingDisabled.ValueBool())
	} else {
		pathForwardingDisabled = nil
	}
	rateLimit := new(int)
	rateLimitPeriod := new(hookdeck.DestinationCreateRequestRateLimitPeriod)
	if m.RateLimit != nil {
		*rateLimit = int(m.RateLimit.Limit.ValueInt64())
		*rateLimitPeriod = hookdeck.DestinationCreateRequestRateLimitPeriod(m.RateLimit.Period.ValueString())
	} else {
		rateLimit = nil
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
	rateLimitPeriod := new(hookdeck.DestinationUpdateRequestRateLimitPeriod)
	if m.RateLimit != nil {
		*rateLimit = int(m.RateLimit.Limit.ValueInt64())
		*rateLimitPeriod = hookdeck.DestinationUpdateRequestRateLimitPeriod(m.RateLimit.Period.ValueString())
	} else {
		rateLimit = nil
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

	var destinationAuthMethodConfig *hookdeck.DestinationAuthMethodConfig
	for _, method := range authenticationMethods {
		if destinationAuthMethodConfig == nil {
			destinationAuthMethodConfig = method.toPayload(m.AuthMethod)
		}
	}

	return destinationAuthMethodConfig
}

func (m *destinationResourceModel) refreshAuthMethod(destination *hookdeck.Destination) {
	if destination.AuthMethod == nil {
		return
	}

	// If users are utilizing a custom JSON payload, the provider should not touch the Terraform state
	// or it will cause conflicts between state & Terraform code.
	if m.AuthMethod != nil && m.AuthMethod.JSON != nil {
		return
	}

	m.AuthMethod = &destinationAuthMethodConfig{}

	for _, method := range authenticationMethods {
		method.refresh(m, destination)
	}
}
