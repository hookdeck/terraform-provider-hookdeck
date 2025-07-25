package connection

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ resource.ResourceWithUpgradeState = &connectionResource{}

// UpgradeState handles the migration from schema version 0 (Set) to version 1 (List)
func (r *connectionResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Migrate from version 0 (SetNestedAttribute) to version 1 (ListNestedAttribute)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: schemaAttributesV0(),
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Read the current state
				var oldState connectionResourceModel
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Sort rules by type priority: Transform > Filter > Retry > Delay
				if len(oldState.Rules) > 0 {
					sort.Slice(oldState.Rules, func(i, j int) bool {
						iPriority := getRulePriority(&oldState.Rules[i])
						jPriority := getRulePriority(&oldState.Rules[j])
						return iPriority < jPriority
					})
				}

				// Set the upgraded state
				resp.Diagnostics.Append(resp.State.Set(ctx, &oldState)...)

				// Add a warning to inform the user
				if len(oldState.Rules) > 0 {
					resp.Diagnostics.AddWarning(
						"Connection Rules Migrated",
						"Connection rules have been migrated from an unordered set to an ordered list. "+
							"Transform rules now appear before filter rules by default. "+
							"You can reorder rules in your configuration as needed.",
					)
				}
			},
		},
	}
}

// getRulePriority returns the priority for sorting rules during migration: Transform > Filter > Retry > Delay
func getRulePriority(r *rule) int {
	if r.TransformRule != nil {
		return 1
	}
	if r.FilterRule != nil {
		return 2
	}
	if r.RetryRule != nil {
		return 3
	}
	if r.DelayRule != nil {
		return 4
	}
	return 5 // Unknown rule type
}
