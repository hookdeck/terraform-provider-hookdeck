package transformation

import (
	"context"

	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transformationResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Code      types.String `tfsdk:"code"`
	CreatedAt types.String `tfsdk:"created_at"`
	ENV       types.String `tfsdk:"env"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// legacyTransformationResourceModel is the model of the v2 name, which also has team_id.
type legacyTransformationResourceModel struct {
	transformationResourceModel
	TeamID types.String `tfsdk:"team_id"`
}

// getModel reads the model from a plan, state or configuration. On the v2
// name it also returns team_id, which records the project in state written
// before project_id existed.
func getModel(ctx context.Context, from shared.ModelSource, legacy bool) (*transformationResourceModel, types.String, diag.Diagnostics) {
	if legacy {
		var m legacyTransformationResourceModel
		diags := from.Get(ctx, &m)
		return &m.transformationResourceModel, m.TeamID, diags
	}
	var m transformationResourceModel
	diags := from.Get(ctx, &m)
	return &m, types.StringNull(), diags
}

// setModel writes the model to state, with team_id on the v2 name.
func setModel(ctx context.Context, to shared.ModelTarget, legacy bool, m *transformationResourceModel) diag.Diagnostics {
	if legacy {
		return to.Set(ctx, &legacyTransformationResourceModel{transformationResourceModel: *m, TeamID: m.ProjectID})
	}
	return to.Set(ctx, m)
}
