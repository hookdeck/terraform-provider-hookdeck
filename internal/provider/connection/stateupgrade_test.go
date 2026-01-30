package connection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestStateUpgradeV0ToV1_RetryRuleOnly replicates issue #191 exactly:
// - V0 state with a connection that has only a retry rule
// - No deduplicate_rule in V0 schema
// - Upgrading should NOT produce "Struct defines fields not found in object: deduplicate_rule"
// See: https://github.com/hookdeck/terraform-provider-hookdeck/issues/191
func TestStateUpgradeV0ToV1_RetryRuleOnly(t *testing.T) {
	// Create mock V0 state matching the user's config from issue #191:
	// rules = [
	//   {
	//     retry_rule = {
	//       count                 = 6
	//       interval              = 60000
	//       strategy              = "exponential"
	//       response_status_codes = ["500-599"]  // Note: this field didn't exist in V0
	//     }
	//   }
	// ]
	v0State := connectionResourceModelV0{
		ID:            types.StringValue("conn_123"),
		Name:          types.StringValue("test-connection"),
		SourceID:      types.StringValue("src_456"),
		DestinationID: types.StringValue("dest_789"),
		CreatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		TeamID:        types.StringValue("team_abc"),
		Description:   types.StringNull(),
		DisabledAt:    types.StringNull(),
		PausedAt:      types.StringNull(),
		Rules: []ruleV0{
			{
				RetryRule: &retryRuleV0{
					Count:    types.Int64Value(6),
					Interval: types.Int64Value(60000),
					Strategy: types.StringValue("exponential"),
					// V0 did NOT have response_status_codes
				},
			},
		},
	}

	// Convert to current model - this is where the bug occurred in issue #191
	result := convertV0ToCurrent(v0State)

	// Verify conversion succeeded
	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result.Rules))
	}

	if result.Rules[0].RetryRule == nil {
		t.Error("expected RetryRule to be set, got nil")
	}

	if result.Rules[0].DeduplicateRule != nil {
		t.Error("expected DeduplicateRule to be nil (didn't exist in V0)")
	}

	if result.Rules[0].RetryRule.Strategy.ValueString() != "exponential" {
		t.Errorf("expected strategy 'exponential', got '%s'", result.Rules[0].RetryRule.Strategy.ValueString())
	}

	if result.Rules[0].RetryRule.Count.ValueInt64() != 6 {
		t.Errorf("expected count 6, got %d", result.Rules[0].RetryRule.Count.ValueInt64())
	}

	if result.Rules[0].RetryRule.Interval.ValueInt64() != 60000 {
		t.Errorf("expected interval 60000, got %d", result.Rules[0].RetryRule.Interval.ValueInt64())
	}

	if !result.Rules[0].RetryRule.ResponseStatusCodes.IsNull() {
		t.Error("expected ResponseStatusCodes to be null (V0 didn't have this field)")
	}

	// Verify other fields are preserved
	if result.ID.ValueString() != "conn_123" {
		t.Errorf("expected ID 'conn_123', got '%s'", result.ID.ValueString())
	}

	if result.Name.ValueString() != "test-connection" {
		t.Errorf("expected Name 'test-connection', got '%s'", result.Name.ValueString())
	}
}

// TestStateUpgradeV0ToV1_MultipleRules tests state upgrade with multiple rule types.
func TestStateUpgradeV0ToV1_MultipleRules(t *testing.T) {
	v0State := connectionResourceModelV0{
		ID:            types.StringValue("conn_multi"),
		Name:          types.StringValue("multi-rule-connection"),
		SourceID:      types.StringValue("src_123"),
		DestinationID: types.StringValue("dest_456"),
		CreatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		TeamID:        types.StringValue("team_xyz"),
		Description:   types.StringNull(),
		DisabledAt:    types.StringNull(),
		PausedAt:      types.StringNull(),
		Rules: []ruleV0{
			{
				FilterRule: &filterRuleV0{
					Headers: &filterRulePropertyV0{
						JSON:    types.StringValue(`{"x-api-key":"secret"}`),
						Boolean: types.BoolNull(),
						Number:  types.NumberNull(),
						String:  types.StringNull(),
					},
				},
			},
			{
				RetryRule: &retryRuleV0{
					Count:    types.Int64Value(5),
					Interval: types.Int64Value(1000),
					Strategy: types.StringValue("linear"),
				},
			},
			{
				DelayRule: &delayRule{
					Delay: types.Int64Value(5000),
				},
			},
		},
	}

	result := convertV0ToCurrent(v0State)

	if len(result.Rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(result.Rules))
	}

	// Verify filter rule conversion (including JSON type conversion)
	if result.Rules[0].FilterRule == nil {
		t.Error("expected FilterRule to be set")
	} else if result.Rules[0].FilterRule.Headers == nil {
		t.Error("expected FilterRule.Headers to be set")
	} else if result.Rules[0].FilterRule.Headers.JSON.ValueString() != `{"x-api-key":"secret"}` {
		t.Errorf("expected JSON '%s', got '%s'", `{"x-api-key":"secret"}`, result.Rules[0].FilterRule.Headers.JSON.ValueString())
	}

	// Verify retry rule
	if result.Rules[1].RetryRule == nil {
		t.Error("expected RetryRule to be set")
	} else if result.Rules[1].RetryRule.Strategy.ValueString() != "linear" {
		t.Errorf("expected strategy 'linear', got '%s'", result.Rules[1].RetryRule.Strategy.ValueString())
	}

	// Verify delay rule
	if result.Rules[2].DelayRule == nil {
		t.Error("expected DelayRule to be set")
	} else if result.Rules[2].DelayRule.Delay.ValueInt64() != 5000 {
		t.Errorf("expected delay 5000, got %d", result.Rules[2].DelayRule.Delay.ValueInt64())
	}

	// Verify no DeduplicateRule on any rule (didn't exist in V0)
	for i, r := range result.Rules {
		if r.DeduplicateRule != nil {
			t.Errorf("expected DeduplicateRule to be nil for rule %d", i)
		}
	}
}

// TestStateUpgradeV0ToV1_NoRules tests state upgrade with no rules.
func TestStateUpgradeV0ToV1_NoRules(t *testing.T) {
	v0State := connectionResourceModelV0{
		ID:            types.StringValue("conn_norules"),
		Name:          types.StringValue("no-rules-connection"),
		SourceID:      types.StringValue("src_789"),
		DestinationID: types.StringValue("dest_012"),
		CreatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		TeamID:        types.StringValue("team_abc"),
		Description:   types.StringNull(),
		DisabledAt:    types.StringNull(),
		PausedAt:      types.StringNull(),
		Rules:         nil,
	}

	result := convertV0ToCurrent(v0State)

	if result.Rules != nil {
		t.Errorf("expected nil rules, got %v", result.Rules)
	}

	// Verify other fields are preserved
	if result.ID.ValueString() != "conn_norules" {
		t.Errorf("expected ID 'conn_norules', got '%s'", result.ID.ValueString())
	}
}

// TestStateUpgradeV0ToV1_EmptyRules tests state upgrade with empty rules slice.
func TestStateUpgradeV0ToV1_EmptyRules(t *testing.T) {
	v0State := connectionResourceModelV0{
		ID:            types.StringValue("conn_emptyrules"),
		Name:          types.StringValue("empty-rules-connection"),
		SourceID:      types.StringValue("src_aaa"),
		DestinationID: types.StringValue("dest_bbb"),
		CreatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		TeamID:        types.StringValue("team_ccc"),
		Description:   types.StringNull(),
		DisabledAt:    types.StringNull(),
		PausedAt:      types.StringNull(),
		Rules:         []ruleV0{},
	}

	result := convertV0ToCurrent(v0State)

	if len(result.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(result.Rules))
	}
}

// TestStateUpgradeV0ToV1_TransformRule tests state upgrade with transform rule.
func TestStateUpgradeV0ToV1_TransformRule(t *testing.T) {
	v0State := connectionResourceModelV0{
		ID:            types.StringValue("conn_transform"),
		Name:          types.StringValue("transform-connection"),
		SourceID:      types.StringValue("src_t1"),
		DestinationID: types.StringValue("dest_t1"),
		CreatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		UpdatedAt:     types.StringValue("2024-01-01T00:00:00Z"),
		TeamID:        types.StringValue("team_t1"),
		Description:   types.StringNull(),
		DisabledAt:    types.StringNull(),
		PausedAt:      types.StringNull(),
		Rules: []ruleV0{
			{
				TransformRule: &transformRule{
					TransformationID: types.StringValue("trfm_123"),
				},
			},
		},
	}

	result := convertV0ToCurrent(v0State)

	if len(result.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(result.Rules))
	}

	if result.Rules[0].TransformRule == nil {
		t.Error("expected TransformRule to be set")
	} else if result.Rules[0].TransformRule.TransformationID.ValueString() != "trfm_123" {
		t.Errorf("expected TransformationID 'trfm_123', got '%s'", result.Rules[0].TransformRule.TransformationID.ValueString())
	}
}

// TestConvertFilterRulePropertyV0ToCurrent tests the filter rule property conversion.
func TestConvertFilterRulePropertyV0ToCurrent(t *testing.T) {
	tests := []struct {
		name     string
		input    *filterRulePropertyV0
		expected *filterRuleProperty
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "JSON property",
			input: &filterRulePropertyV0{
				JSON:    types.StringValue(`{"key":"value"}`),
				Boolean: types.BoolNull(),
				Number:  types.NumberNull(),
				String:  types.StringNull(),
			},
			expected: &filterRuleProperty{
				JSON:    jsontypes.NewNormalizedValue(`{"key":"value"}`),
				Boolean: types.BoolNull(),
				Number:  types.NumberNull(),
				String:  types.StringNull(),
			},
		},
		{
			name: "Boolean property",
			input: &filterRulePropertyV0{
				JSON:    types.StringNull(),
				Boolean: types.BoolValue(true),
				Number:  types.NumberNull(),
				String:  types.StringNull(),
			},
			expected: &filterRuleProperty{
				JSON:    jsontypes.NewNormalizedNull(),
				Boolean: types.BoolValue(true),
				Number:  types.NumberNull(),
				String:  types.StringNull(),
			},
		},
		{
			name: "String property",
			input: &filterRulePropertyV0{
				JSON:    types.StringNull(),
				Boolean: types.BoolNull(),
				Number:  types.NumberNull(),
				String:  types.StringValue("test-string"),
			},
			expected: &filterRuleProperty{
				JSON:    jsontypes.NewNormalizedNull(),
				Boolean: types.BoolNull(),
				Number:  types.NumberNull(),
				String:  types.StringValue("test-string"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFilterRulePropertyV0ToCurrent(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Error("expected nil result")
				}
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			// Compare JSON values
			if tt.expected.JSON.IsNull() != result.JSON.IsNull() {
				t.Errorf("JSON null mismatch: expected %v, got %v", tt.expected.JSON.IsNull(), result.JSON.IsNull())
			} else if !tt.expected.JSON.IsNull() && tt.expected.JSON.ValueString() != result.JSON.ValueString() {
				t.Errorf("JSON value mismatch: expected %s, got %s", tt.expected.JSON.ValueString(), result.JSON.ValueString())
			}

			// Compare Boolean values
			if tt.expected.Boolean.IsNull() != result.Boolean.IsNull() {
				t.Errorf("Boolean null mismatch: expected %v, got %v", tt.expected.Boolean.IsNull(), result.Boolean.IsNull())
			} else if !tt.expected.Boolean.IsNull() && tt.expected.Boolean.ValueBool() != result.Boolean.ValueBool() {
				t.Errorf("Boolean value mismatch: expected %v, got %v", tt.expected.Boolean.ValueBool(), result.Boolean.ValueBool())
			}

			// Compare String values
			if tt.expected.String.IsNull() != result.String.IsNull() {
				t.Errorf("String null mismatch: expected %v, got %v", tt.expected.String.IsNull(), result.String.IsNull())
			} else if !tt.expected.String.IsNull() && tt.expected.String.ValueString() != result.String.ValueString() {
				t.Errorf("String value mismatch: expected %s, got %s", tt.expected.String.ValueString(), result.String.ValueString())
			}
		})
	}
}

// TestGetRulePriority tests the rule priority function includes DeduplicateRule.
func TestGetRulePriority(t *testing.T) {
	tests := []struct {
		name             string
		rule             rule
		expectedPriority int
	}{
		{
			name:             "TransformRule has highest priority",
			rule:             rule{TransformRule: &transformRule{TransformationID: types.StringValue("t1")}},
			expectedPriority: 1,
		},
		{
			name:             "FilterRule has second priority",
			rule:             rule{FilterRule: &filterRule{}},
			expectedPriority: 2,
		},
		{
			name:             "DeduplicateRule has third priority",
			rule:             rule{DeduplicateRule: &deduplicateRule{Window: types.Int64Value(1000)}},
			expectedPriority: 3,
		},
		{
			name:             "DelayRule has fourth priority",
			rule:             rule{DelayRule: &delayRule{Delay: types.Int64Value(1000)}},
			expectedPriority: 4,
		},
		{
			name:             "RetryRule has fifth priority",
			rule:             rule{RetryRule: &retryRule{Strategy: types.StringValue("linear")}},
			expectedPriority: 5,
		},
		{
			name:             "Empty rule has lowest priority",
			rule:             rule{},
			expectedPriority: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := getRulePriority(&tt.rule)
			if priority != tt.expectedPriority {
				t.Errorf("expected priority %d, got %d", tt.expectedPriority, priority)
			}
		})
	}
}
