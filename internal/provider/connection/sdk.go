package connection

import (
	"encoding/json"
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

func (m *connectionResourceModel) ToCreatePayload() *hookdeck.ConnectionCreateRequest {
	return &hookdeck.ConnectionCreateRequest{
		Name:          hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
		Description:   hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		DestinationId: hookdeck.OptionalOrNull(m.DestinationID.ValueStringPointer()),
		Rules:         hookdeck.OptionalOrNull(m.getRules()),
		SourceId:      hookdeck.OptionalOrNull(m.SourceID.ValueStringPointer()),
	}
}

func (m *connectionResourceModel) ToUpdatePayload() *hookdeck.ConnectionUpdateRequest {
	return &hookdeck.ConnectionUpdateRequest{
		Name:        hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Rules:       hookdeck.OptionalOrNull(m.getRules()),
	}
}

func (m *connectionResourceModel) getRules() *[]*hookdeck.Rule {
	var rules []*hookdeck.Rule = []*hookdeck.Rule{}

	for _, ruleItem := range m.Rules {
		if ruleItem.DelayRule != nil {
			delayRule := hookdeck.DelayRule{
				Delay: int(ruleItem.DelayRule.Delay.ValueInt64()),
			}
			rules = append(rules, hookdeck.NewRuleFromDelayRule(&delayRule))
		}
		if ruleItem.FilterRule != nil {
			filterRule := hookdeck.FilterRule{
				Body:    transformFilterRuleProperty(ruleItem.FilterRule.Body),
				Headers: transformFilterRuleProperty(ruleItem.FilterRule.Headers),
				Path:    transformFilterRuleProperty(ruleItem.FilterRule.Path),
				Query:   transformFilterRuleProperty(ruleItem.FilterRule.Query),
			}
			rules = append(rules, hookdeck.NewRuleFromFilterRule(&filterRule))
		}
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
			transformRule := hookdeck.NewTransformRuleFromTransformReference(&hookdeck.TransformReference{
				TransformationId: ruleItem.TransformRule.TransformationID.ValueString(),
			})
			rules = append(rules, hookdeck.NewRuleFromTransformRule(transformRule))
		}
	}

	return &rules
}

func transformFilterRuleProperty(property *filterRuleProperty) *hookdeck.FilterRuleProperty {
	if !property.Boolean.IsUnknown() && !property.Boolean.IsNull() {
		return hookdeck.NewFilterRulePropertyFromBooleanOptional(property.Boolean.ValueBoolPointer())
	}
	if !property.JSON.IsUnknown() && !property.JSON.IsNull() {
		// parse string to JSON
		var jsonData map[string]any
		jsonBytes := []byte(property.JSON.ValueString())
		json.Unmarshal(jsonBytes, &jsonData)
		return hookdeck.NewFilterRulePropertyFromStringUnknownMapOptional(jsonData)
	}
	if !property.Number.IsUnknown() && !property.Number.IsNull() {
		return hookdeck.NewFilterRulePropertyFromDoubleOptional(property.Number.ValueFloat64Pointer())
	}
	if !property.String.IsUnknown() && !property.String.IsNull() {
		return hookdeck.NewFilterRulePropertyFromStringOptional(property.String.ValueStringPointer())
	}
	return nil
}
