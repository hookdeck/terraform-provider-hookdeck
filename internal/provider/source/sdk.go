package source

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *sourceResourceModel) Refresh(source *hookdeck.Source) {
	allowedHTTPMethods := []string{}
	for _, v := range *source.AllowedHttpMethods {
		allowedHTTPMethods = append(allowedHTTPMethods, string(v))
	}
	var diags diag.Diagnostics
	m.AllowedHTTPMethods, diags = types.ListValueFrom(context.Background(), types.StringType, allowedHTTPMethods)
	if diags.HasError() {
		panic(diags)
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
	if source.DisabledAt != nil {
		m.DisabledAt = types.StringValue(source.DisabledAt.Format(time.RFC3339))
	} else {
		m.DisabledAt = types.StringNull()
	}
	m.ID = types.StringValue(source.Id)
	m.Name = types.StringValue(source.Name)
	m.TeamID = types.StringValue(source.TeamId)
	m.UpdatedAt = types.StringValue(source.UpdatedAt.Format(time.RFC3339))
	m.URL = types.StringValue(source.Url)
}

func (m *sourceResourceModel) ToCreatePayload() *hookdeck.SourceCreateRequest {
	return &hookdeck.SourceCreateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        m.Name.ValueString(),
	}
}

func (m *sourceResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
	}
}
