package destination

import (
	"context"

	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationResourceModel struct {
	ProjectID   types.String         `tfsdk:"project_id"`
	Config      jsontypes.Normalized `tfsdk:"config"`
	CreatedAt   types.String         `tfsdk:"created_at"`
	Description types.String         `tfsdk:"description"`
	DisabledAt  types.String         `tfsdk:"disabled_at"`
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Type        types.String         `tfsdk:"type"`
	UpdatedAt   types.String         `tfsdk:"updated_at"`
}

// legacyDestinationResourceModel is the model of the v2 name, which also has team_id.
type legacyDestinationResourceModel struct {
	destinationResourceModel
	TeamID types.String `tfsdk:"team_id"`
}

// getModel reads the model from a plan, state or configuration. On the v2
// name it also returns team_id, which records the project in state written
// before project_id existed.
func getModel(ctx context.Context, from shared.ModelSource, legacy bool) (*destinationResourceModel, types.String, diag.Diagnostics) {
	if legacy {
		var m legacyDestinationResourceModel
		diags := from.Get(ctx, &m)
		return &m.destinationResourceModel, m.TeamID, diags
	}
	var m destinationResourceModel
	diags := from.Get(ctx, &m)
	return &m, types.StringNull(), diags
}

// setModel writes the model to state, with team_id on the v2 name.
func setModel(ctx context.Context, to shared.ModelTarget, legacy bool, m *destinationResourceModel) diag.Diagnostics {
	if legacy {
		return to.Set(ctx, &legacyDestinationResourceModel{destinationResourceModel: *m, TeamID: m.ProjectID})
	}
	return to.Set(ctx, m)
}
