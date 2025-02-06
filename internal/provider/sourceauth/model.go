package sourceauth

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceAuthResourceModel struct {
	AuthType types.String         `tfsdk:"auth_type"`
	Auth     jsontypes.Normalized `tfsdk:"auth"`
	SourceID types.String         `tfsdk:"source_id"`
}
