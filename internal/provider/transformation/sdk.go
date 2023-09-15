package transformation

import (
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *transformationResourceModel) Refresh(transformation *hookdeck.Transformation) {
	m.Code = types.StringValue(transformation.Code)
	m.CreatedAt = types.StringValue(transformation.CreatedAt.Format(time.RFC3339))
	if transformation.Env != nil && len(transformation.Env) > 0 {
		envBytes, err := json.Marshal(transformation.Env)
		if err == nil {
			m.ENV = types.StringValue(string(envBytes))
		} else {
			m.ENV = types.StringNull()
		}
	} else {
		m.ENV = types.StringNull()
	}
	m.ID = types.StringValue(transformation.Id)
	m.Name = types.StringValue(transformation.Name)
	m.TeamID = types.StringValue(transformation.TeamId)
	m.UpdatedAt = types.StringValue(transformation.UpdatedAt.Format(time.RFC3339))
}

func (m *transformationResourceModel) ToCreatePayload() *hookdeck.TransformationCreateRequest {
	env := m.getENV()

	return &hookdeck.TransformationCreateRequest{
		Code: m.Code.ValueString(),
		Env:  hookdeck.OptionalOrNull(&env),
		Name: m.Name.ValueString(),
	}
}

func (m *transformationResourceModel) ToUpdatePayload() *hookdeck.TransformationUpdateRequest {
	env := m.getENV()

	return &hookdeck.TransformationUpdateRequest{
		Code: hookdeck.OptionalOrNull(m.Code.ValueStringPointer()),
		Env:  hookdeck.OptionalOrNull(&env),
		Name: hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
	}
}

func (m *transformationResourceModel) getENV() map[string]string {
	var envData map[string]string = nil
	if !m.ENV.IsUnknown() && !m.ENV.IsNull() {
		envBytes := []byte(m.ENV.ValueString())
		if err := json.Unmarshal(envBytes, &envData); err != nil {
			return nil
		}
	}
	return envData
}
