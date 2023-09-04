package transformation

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *transformationResourceModel) Refresh(transformation *hookdeck.Transformation) {
	m.Code = types.StringValue(transformation.Code)
	m.CreatedAt = types.StringValue(transformation.CreatedAt.Format(time.RFC3339))
	// if transformation.Env != nil || len(transformation.Env) != 0 {
	// 	m.ENV = types.StringValue(fmt.Sprint(transformation.Env))
	// } else {
	// 	m.ENV = types.StringNull()
	// }
	m.ID = types.StringValue(transformation.Id)
	m.Name = types.StringValue(transformation.Name)
	m.TeamID = types.StringValue(transformation.TeamId)
	m.UpdatedAt = types.StringValue(transformation.UpdatedAt.Format(time.RFC3339))
}

func (m *transformationResourceModel) ToCreatePayload() *hookdeck.TransformationCreateRequest {
	// var envData map[string]string = nil
	// if !m.ENV.IsUnknown() && !m.ENV.IsNull() {
	// 	envBytes := []byte(m.ENV.ValueString())
	// 	json.Unmarshal(envBytes, &envData)
	// }

	return &hookdeck.TransformationCreateRequest{
		Code: m.Code.ValueString(),
		// Env:  hookdeck.OptionalOrNull(&envData),
		Env:  nil,
		Name: m.Name.ValueString(),
	}
}

func (m *transformationResourceModel) ToUpdatePayload() *hookdeck.TransformationUpdateRequest {
	return &hookdeck.TransformationUpdateRequest{
		Code: hookdeck.OptionalOrNull(m.Code.ValueStringPointer()),
		Name: hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
	}
}
