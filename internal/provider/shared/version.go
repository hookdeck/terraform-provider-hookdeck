package shared

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// ConfigSchemaVersion is the schema version of sources and destinations.
// Version 1 has the attributes of version 0; it marks state whose config holds
// everything Hookdeck returns. Terraform refuses to hand such state to a
// provider that knows only version 0, which would read every key in it as set
// by the configuration and clear the ones the configuration does not have.
const ConfigSchemaVersion int64 = 1

// UnchangedStateUpgraders accepts state of schema version 0 as it is.
func UnchangedStateUpgraders(current resourceschema.Schema) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &current,
			StateUpgrader: func(_ context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				if req.State == nil {
					resp.Diagnostics.AddError("Unsupported state", "The stored state could not be decoded. Please report this issue to the provider developers.")
					return
				}
				resp.State.Raw = req.State.Raw
			},
		},
	}
}
