package connection

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *connectionResourceModel) Refresh(connection *hookdeck.Connection) {
	if connection.ArchivedAt != nil {
		m.ArchivedAt = types.StringValue(connection.ArchivedAt.Format(time.RFC3339))
	} else {
		m.ArchivedAt = types.StringNull()
	}
	m.CreatedAt = types.StringValue(connection.CreatedAt.Format(time.RFC3339))
	m.DestinationID = types.StringValue(connection.Destination.Id)
	m.ID = types.StringValue(connection.Id)
	if connection.Name != nil {
		m.Name = types.StringValue(*connection.Name)
	} else {
		m.Name = types.StringNull()
	}
	if connection.PausedAt != nil {
		m.PausedAt = types.StringValue(connection.PausedAt.Format(time.RFC3339))
	} else {
		m.PausedAt = types.StringNull()
	}
	m.SourceID = types.StringValue(connection.Source.Id)
	m.TeamID = types.StringValue(connection.TeamId)
	m.UpdatedAt = types.StringValue(connection.UpdatedAt.Format(time.RFC3339))
}

func (m *connectionResourceModel) ToCreatePayload() *hookdeck.CreateConnectionRequest {
	return &hookdeck.CreateConnectionRequest{
		Name:          m.Name.ValueStringPointer(),
		Description:   m.Description.ValueStringPointer(),
		DestinationId: m.DestinationID.ValueStringPointer(),
		Rules:         m.getRules(),
		SourceId:      m.SourceID.ValueStringPointer(),
	}
}

func (m *connectionResourceModel) ToUpdatePayload() *hookdeck.UpdateConnectionRequest {
	return &hookdeck.UpdateConnectionRequest{
		Name:        m.Name.ValueStringPointer(),
		Description: m.Description.ValueStringPointer(),
		Rules:       m.getRules(),
	}
}

func (m *connectionResourceModel) getRules() []*hookdeck.Rule {
	var rules []*hookdeck.Rule = nil

	for _, ruleItem := range m.Rules {
		if ruleItem.DelayRule != nil {
			delayRule := hookdeck.DelayRule{
				Delay: int(ruleItem.DelayRule.Delay.ValueInt64()),
			}
			rules = append(rules, hookdeck.NewRuleFromDelayRule(&delayRule))
		}
		// if ruleItem.FilterRule != nil {
		// 	filterRule := hookdeck.FilterRule {
		// 	}
		// 	rules = append(rules, hookdeck.NewRuleFromFilterRule(&filterRule))
		// }
		if ruleItem.RetryRule != nil {
			count := new(int)
			if !ruleItem.RetryRule.Count.IsUnknown() && !ruleItem.RetryRule.Count.IsNull() {
				*count = int(ruleItem.RetryRule.Count.ValueInt64())
			} else {
				count = nil
			}
			interval := new(int)
			if !ruleItem.RetryRule.Interval.IsUnknown() && !ruleItem.RetryRule.Interval.IsNull() {
				*interval = int(ruleItem.RetryRule.Interval.ValueInt64())
			} else {
				interval = nil
			}
			retryRule := hookdeck.RetryRule{
				Strategy: hookdeck.RetryStrategy(ruleItem.RetryRule.Strategy.ValueString()),
				Interval: interval,
				Count:    count,
			}
			rules = append(rules, hookdeck.NewRuleFromRetryRule(&retryRule))
		}
		if ruleItem.TransformRule != nil {
			transformRule := hookdeck.TransformRule{
				TransformReference: &hookdeck.TransformReference{
					TransformationId: ruleItem.TransformRule.TransformationID.ValueString(),
				},
			}
			rules = append(rules, hookdeck.NewRuleFromTransformRule(&transformRule))
		}
	}

	return rules
}
