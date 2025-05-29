package connection

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *connectionResourceModel) Refresh(connection *hookdeck.Connection) {
	m.CreatedAt = types.StringValue(connection.CreatedAt.Format(time.RFC3339))
	m.DestinationID = types.StringValue(connection.Destination.Id)
	if connection.DisabledAt != nil {
		m.DisabledAt = types.StringValue(connection.DisabledAt.Format(time.RFC3339))
	} else {
		m.DisabledAt = types.StringNull()
	}
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
	if len(connection.Rules) > 0 {
		m.Rules = refreshRules(connection)
	}
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
	//nolint:staticcheck // keep explicit type for readability
	var rules []*hookdeck.Rule = []*hookdeck.Rule{}

	for _, ruleItem := range m.Rules {
		if ruleItem.DelayRule != nil {
			delayRule := hookdeck.DelayRule{
				Delay: int(ruleItem.DelayRule.Delay.ValueInt64()),
			}
			rules = append(rules, hookdeck.NewRuleFromDelay(&delayRule))
		}
		if ruleItem.FilterRule != nil {
			filterRule := hookdeck.FilterRule{
				Body:    transformFilterRuleProperty(ruleItem.FilterRule.Body),
				Headers: transformFilterRuleProperty(ruleItem.FilterRule.Headers),
				Path:    transformFilterRuleProperty(ruleItem.FilterRule.Path),
				Query:   transformFilterRuleProperty(ruleItem.FilterRule.Query),
			}
			rules = append(rules, hookdeck.NewRuleFromFilter(&filterRule))
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
			rules = append(rules, hookdeck.NewRuleFromRetry(&retryRule))
		}
		if ruleItem.TransformRule != nil {
			transformRule := hookdeck.TransformRule{
				TransformationId: ruleItem.TransformRule.TransformationID.ValueStringPointer(),
			}
			rules = append(rules, hookdeck.NewRuleFromTransform(&transformRule))
		}
	}

	return &rules
}

func transformFilterRuleProperty(property *filterRuleProperty) *hookdeck.FilterRuleProperty {
	if property == nil {
		return nil
	}
	if !property.Boolean.IsUnknown() && !property.Boolean.IsNull() {
		return hookdeck.NewFilterRulePropertyFromBooleanOptional(property.Boolean.ValueBoolPointer())
	}
	if !property.JSON.IsUnknown() && !property.JSON.IsNull() {
		// parse string to JSON
		var jsonData map[string]any
		jsonBytes := []byte(property.JSON.ValueString())
		if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
			return nil
		}
		return hookdeck.NewFilterRulePropertyFromStringUnknownMapOptional(jsonData)
	}
	if !property.Number.IsUnknown() && !property.Number.IsNull() {
		float, _ := property.Number.ValueBigFloat().Float64()
		return hookdeck.NewFilterRulePropertyFromDoubleOptional(&float)
	}
	if !property.String.IsUnknown() && !property.String.IsNull() {
		return hookdeck.NewFilterRulePropertyFromStringOptional(property.String.ValueStringPointer())
	}
	return nil
}

func refreshRules(connection *hookdeck.Connection) []rule {
	//nolint:staticcheck
	var rules []rule = []rule{}

	for _, ruleItem := range connection.Rules {
		if ruleItem.Delay != nil {
			delayRule := delayRule{
				Delay: types.Int64Value(int64(ruleItem.Delay.Delay)),
			}
			rules = append(rules, rule{DelayRule: &delayRule})
		}
		if ruleItem.Filter != nil {
			filterRule := filterRule{}
			if ruleItem.Filter.Body != nil {
				filterRule.Body = refreshFilterRuleProperty(ruleItem.Filter.Body)
			}
			if ruleItem.Filter.Headers != nil {
				filterRule.Headers = refreshFilterRuleProperty(ruleItem.Filter.Headers)
			}
			if ruleItem.Filter.Path != nil {
				filterRule.Path = refreshFilterRuleProperty(ruleItem.Filter.Path)
			}
			if ruleItem.Filter.Query != nil {
				filterRule.Query = refreshFilterRuleProperty(ruleItem.Filter.Query)
			}
			rules = append(rules, rule{FilterRule: &filterRule})
		}
		if ruleItem.Retry != nil {
			retryRule := retryRule{}
			if ruleItem.Retry.Count != nil {
				retryRule.Count = types.Int64Value(int64(*ruleItem.Retry.Count))
			}
			if ruleItem.Retry.Interval != nil {
				retryRule.Interval = types.Int64Value(int64(*ruleItem.Retry.Interval))
			}
			retryRule.Strategy = types.StringValue((string)(ruleItem.Retry.Strategy))
			rules = append(rules, rule{RetryRule: &retryRule})
		}
		if ruleItem.Transform != nil {
			transformRule := transformRule{
				// in this case (refresh), the ID should always exist
				// so we don't have to worry the ruleItem.Transform.TransformationId is nil
				TransformationID: types.StringValue(*ruleItem.Transform.TransformationId),
			}
			rules = append(rules, rule{TransformRule: &transformRule})
		}
	}

	return rules
}

func refreshFilterRuleProperty(property *hookdeck.FilterRuleProperty) *filterRuleProperty {
	if property == nil {
		return nil
	}

	var filterRuleProperty = filterRuleProperty{}

	if property.BooleanOptional != nil {
		filterRuleProperty.Boolean = types.BoolValue(*property.BooleanOptional)
	}
	if property.DoubleOptional != nil {
		number := new(big.Float)
		number.SetFloat64(*property.DoubleOptional)
		filterRuleProperty.Number = types.NumberValue(number)
	}
	if property.StringOptional != nil {
		filterRuleProperty.String = types.StringValue(*property.StringOptional)
	}
	if property.StringUnknownMapOptional != nil {
		marshalledJSON, _ := json.Marshal(property.StringUnknownMapOptional)
		filterRuleProperty.JSON = types.StringValue(string(marshalledJSON))
	}

	return &filterRuleProperty
}
