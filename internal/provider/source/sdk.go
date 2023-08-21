package source

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *sourceResourceModel) Refresh(source *hookdeck.Source) {
	m.AllowedHTTPMethods = nil
	for _, v := range *source.AllowedHttpMethods {
		m.AllowedHTTPMethods = append(m.AllowedHTTPMethods, types.StringValue(string(v)))
	}
	if source.ArchivedAt != nil {
		m.ArchivedAt = types.StringValue(source.ArchivedAt.Format(time.RFC3339))
	} else {
		m.ArchivedAt = types.StringNull()
	}
	m.CreatedAt = types.StringValue(source.CreatedAt.Format(time.RFC3339))
	if m.CustomResponse == nil {
		m.CustomResponse = &sourceCustomResponse{}
	}
	if source.CustomResponse == nil {
		m.CustomResponse = nil
	} else {
		m.CustomResponse = &sourceCustomResponse{}
		m.CustomResponse.Body = types.StringValue(source.CustomResponse.Body)
		m.CustomResponse.ContentType = types.StringValue(string(source.CustomResponse.ContentType))
	}
	m.ID = types.StringValue(source.Id)
	m.Name = types.StringValue(source.Name)
	m.TeamID = types.StringValue(source.TeamId)
	m.UpdatedAt = types.StringValue(source.UpdatedAt.Format(time.RFC3339))
	m.URL = types.StringValue(source.Url)
}

func (m *sourceResourceModel) ToCreatePayload() *hookdeck.CreateSourceRequest {
	var allowedHTTPMethods []hookdeck.SourceAllowedHttpMethodItem = nil
	for _, allowedHTTPMethodsItem := range m.AllowedHTTPMethods {
		allowedHTTPMethods = append(allowedHTTPMethods, hookdeck.SourceAllowedHttpMethodItem(allowedHTTPMethodsItem.ValueString()))
	}
	var customResponse *hookdeck.SourceCustomResponse
	if m.CustomResponse != nil {
		body := m.CustomResponse.Body.ValueString()
		contentType := hookdeck.SourceCustomResponseContentType(m.CustomResponse.ContentType.ValueString())
		customResponse = &hookdeck.SourceCustomResponse{
			Body:        body,
			ContentType: contentType,
		}
	}
	return &hookdeck.CreateSourceRequest{
		AllowedHttpMethods: &allowedHTTPMethods,
		CustomResponse:     customResponse,
		Description:        m.Description.ValueStringPointer(),
		Name:               m.Name.ValueString(),
	}
}

func (m *sourceResourceModel) ToUpdatePayload() *hookdeck.UpdateSourceRequest {
	var allowedHTTPMethods []hookdeck.SourceAllowedHttpMethodItem = nil
	for _, allowedHTTPMethodsItem := range m.AllowedHTTPMethods {
		allowedHTTPMethods = append(allowedHTTPMethods, hookdeck.SourceAllowedHttpMethodItem(allowedHTTPMethodsItem.ValueString()))
	}
	var customResponse *hookdeck.SourceCustomResponse = nil
	if m.CustomResponse != nil {
		body := m.CustomResponse.Body.ValueString()
		contentType := hookdeck.SourceCustomResponseContentType(m.CustomResponse.ContentType.ValueString())
		customResponse = &hookdeck.SourceCustomResponse{
			Body:        body,
			ContentType: contentType,
		}
	}
	return &hookdeck.UpdateSourceRequest{
		AllowedHttpMethods: &allowedHTTPMethods,
		CustomResponse:     customResponse,
		Description:        m.Description.ValueStringPointer(),
		Name:               m.Name.ValueStringPointer(),
	}
}
