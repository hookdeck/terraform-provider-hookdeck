package sourceverification

import (
	"terraform-provider-hookdeck/internal/provider/sourceverification/generated"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceVerificationResourceModel struct {
	SourceID     types.String                  `tfsdk:"source_id"`
	Verification *generated.SourceVerification `tfsdk:"verification"`
}
