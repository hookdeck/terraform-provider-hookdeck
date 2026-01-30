package connection

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithUpgradeState = &connectionResource{}

// UpgradeState handles the migration from schema version 0 (Set) to version 1 (List).
func (r *connectionResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Migrate from version 0 (SetNestedAttribute) to version 1 (ListNestedAttribute)
		0: {
			PriorSchema: &schema.Schema{
				Attributes: schemaAttributesV0(),
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				// Read the current state using V0 model
				// This fixes issue #191: V0 schema doesn't have deduplicate_rule,
				// so we must use a V0-compatible model struct
				var oldState connectionResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &oldState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Convert V0 model to current model
				newState := convertV0ToCurrent(oldState)

				// Sort rules by type priority: Transform > Filter > Deduplicate > Delay > Retry
				if len(newState.Rules) > 0 {
					sort.Slice(newState.Rules, func(i, j int) bool {
						iPriority := getRulePriority(&newState.Rules[i])
						jPriority := getRulePriority(&newState.Rules[j])
						return iPriority < jPriority
					})
				}

				// Set the upgraded state
				resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

				// Add a warning to inform the user
				if len(newState.Rules) > 0 {
					resp.Diagnostics.AddWarning(
						"Potential Action Required: Connection Rules Ordering",
						"Hookdeck has migrated all existing connection rules to an ordered list format.\n"+
							"Only Transform and Filter rules now execute in the order they appear.\n"+
							"To maintain existing behaviour, the platform has placed Transform rules before Filter rules.\n"+
							"Therefore, Terraform will detect any difference in rule order as a change.\n"+
							"To avoid unwanted changes, ensure your Transform and Filter rules are ordered as intended—"+
							"with Transform rules before Filter rules to match previous behaviour.\n\n"+
							"See https://hkdk.link/tf-v1-v2-migration-guide for more details.",
					)
				}
			},
		},
	}
}

// getRulePriority returns the priority for sorting rules during migration: Transform > Filter > Deduplicate > Delay > Retry.
func getRulePriority(r *rule) int {
	if r.TransformRule != nil {
		return 1
	}
	if r.FilterRule != nil {
		return 2
	}
	if r.DeduplicateRule != nil {
		return 3
	}
	if r.DelayRule != nil {
		return 4
	}
	if r.RetryRule != nil {
		return 5
	}
	return 6 // Unknown rule type
}

// convertV0ToCurrent converts a V0 connection model to the current model.
// This is used during state upgrade from schema version 0 to version 1.
func convertV0ToCurrent(v0 connectionResourceModelV0) connectionResourceModel {
	return connectionResourceModel{
		CreatedAt:     v0.CreatedAt,
		Description:   v0.Description,
		DestinationID: v0.DestinationID,
		DisabledAt:    v0.DisabledAt,
		ID:            v0.ID,
		Name:          v0.Name,
		PausedAt:      v0.PausedAt,
		Rules:         convertRulesV0ToCurrent(v0.Rules),
		SourceID:      v0.SourceID,
		TeamID:        v0.TeamID,
		UpdatedAt:     v0.UpdatedAt,
	}
}

// convertRulesV0ToCurrent converts V0 rules to the current rule format.
func convertRulesV0ToCurrent(rulesV0 []ruleV0) []rule {
	if rulesV0 == nil {
		return nil
	}

	result := make([]rule, 0, len(rulesV0))
	for _, r := range rulesV0 {
		newRule := rule{}

		if r.DelayRule != nil {
			newRule.DelayRule = r.DelayRule
		}
		if r.FilterRule != nil {
			newRule.FilterRule = convertFilterRuleV0ToCurrent(r.FilterRule)
		}
		if r.RetryRule != nil {
			newRule.RetryRule = &retryRule{
				Count:               r.RetryRule.Count,
				Interval:            r.RetryRule.Interval,
				Strategy:            r.RetryRule.Strategy,
				ResponseStatusCodes: types.ListNull(types.StringType), // V0 didn't have this field
			}
		}
		if r.TransformRule != nil {
			newRule.TransformRule = r.TransformRule
		}
		// DeduplicateRule stays nil (didn't exist in V0)

		result = append(result, newRule)
	}
	return result
}

// convertFilterRuleV0ToCurrent converts a V0 filter rule to the current format.
// The main difference is that V0 used plain types.String for JSON, while current uses jsontypes.Normalized.
func convertFilterRuleV0ToCurrent(v0 *filterRuleV0) *filterRule {
	if v0 == nil {
		return nil
	}

	return &filterRule{
		Body:    convertFilterRulePropertyV0ToCurrent(v0.Body),
		Headers: convertFilterRulePropertyV0ToCurrent(v0.Headers),
		Path:    convertFilterRulePropertyV0ToCurrent(v0.Path),
		Query:   convertFilterRulePropertyV0ToCurrent(v0.Query),
	}
}

// convertFilterRulePropertyV0ToCurrent converts a V0 filter rule property to the current format.
func convertFilterRulePropertyV0ToCurrent(v0 *filterRulePropertyV0) *filterRuleProperty {
	if v0 == nil {
		return nil
	}

	result := &filterRuleProperty{
		Boolean: v0.Boolean,
		Number:  v0.Number,
		String:  v0.String,
	}

	// Convert plain string JSON to jsontypes.Normalized
	if !v0.JSON.IsNull() && !v0.JSON.IsUnknown() {
		result.JSON = jsontypes.NewNormalizedValue(v0.JSON.ValueString())
	} else {
		result.JSON = jsontypes.NewNormalizedNull()
	}

	return result
}
