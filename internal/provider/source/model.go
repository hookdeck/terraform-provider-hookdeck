package source

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceResourceModel struct {
	Config      jsontypes.Normalized `tfsdk:"config"`
	CreatedAt   types.String         `tfsdk:"created_at"`
	Description types.String         `tfsdk:"description"`
	DisabledAt  types.String         `tfsdk:"disabled_at"`
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	TeamID      types.String         `tfsdk:"team_id"`
	Type        types.String         `tfsdk:"type"`
	UpdatedAt   types.String         `tfsdk:"updated_at"`
	URL         types.String         `tfsdk:"url"`
}
