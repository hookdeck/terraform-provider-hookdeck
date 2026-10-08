package source

import (
	"context"

	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceResourceModel struct {
	ProjectID   types.String         `tfsdk:"project_id"`
	Config      jsontypes.Normalized `tfsdk:"config"`
	CreatedAt   types.String         `tfsdk:"created_at"`
	Description types.String         `tfsdk:"description"`
	DisabledAt  types.String         `tfsdk:"disabled_at"`
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Type        types.String         `tfsdk:"type"`
	UpdatedAt   types.String         `tfsdk:"updated_at"`
	URL         types.String         `tfsdk:"url"`
}

// legacySourceResourceModel is the model of the v2 name, which also has team_id.
type legacySourceResourceModel struct {
	sourceResourceModel
	TeamID types.String `tfsdk:"team_id"`
}

// getModel reads the model from a plan, state or configuration. On the v2
// name it also returns team_id, which records the project in state written
// before project_id existed.
func getModel(ctx context.Context, from shared.ModelSource, legacy bool) (*sourceResourceModel, types.String, diag.Diagnostics) {
	if legacy {
		var m legacySourceResourceModel
		diags := from.Get(ctx, &m)
		return &m.sourceResourceModel, m.TeamID, diags
	}
	var m sourceResourceModel
	diags := from.Get(ctx, &m)
	return &m, types.StringNull(), diags
}

// setModel writes the model to state, with team_id on the v2 name.
func setModel(ctx context.Context, to shared.ModelTarget, legacy bool, m *sourceResourceModel) diag.Diagnostics {
	if legacy {
		return to.Set(ctx, &legacySourceResourceModel{sourceResourceModel: *m, TeamID: m.ProjectID})
	}
	return to.Set(ctx, m)
}
