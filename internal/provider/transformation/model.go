package transformation

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type transformationResourceModel struct {
	Code      types.String `tfsdk:"code"`
	CreatedAt types.String `tfsdk:"created_at"`
	ENV       types.String `tfsdk:"env"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	TeamID    types.String `tfsdk:"team_id"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
