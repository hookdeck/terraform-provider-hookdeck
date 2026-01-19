package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiVersion = "2025-07-01"

func (m *connectionResourceModel) Refresh(connection map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Required fields
	if createdAt, ok := connection["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(createdAt)
	} else {
		diags.AddError("Error parsing created_at", "Expected string value")
		return diags
	}

	if id, ok := connection["id"].(string); ok {
		m.ID = types.StringValue(id)
	} else {
		diags.AddError("Error parsing id", "Expected string value")
		return diags
	}

	if teamID, ok := connection["team_id"].(string); ok {
		m.TeamID = types.StringValue(teamID)
	} else {
		diags.AddError("Error parsing team_id", "Expected string value")
		return diags
	}

	if updatedAt, ok := connection["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(updatedAt)
	} else {
		diags.AddError("Error parsing updated_at", "Expected string value")
		return diags
	}

	// Handle destination - can be string ID or object
	if destination, ok := connection["destination"]; ok {
		switch v := destination.(type) {
		case string:
			m.DestinationID = types.StringValue(v)
		case map[string]interface{}:
			if id, ok := v["id"].(string); ok {
				m.DestinationID = types.StringValue(id)
			}
		}
	}

	// Handle source - can be string ID or object
	if source, ok := connection["source"]; ok {
		switch v := source.(type) {
		case string:
			m.SourceID = types.StringValue(v)
		case map[string]interface{}:
			if id, ok := v["id"].(string); ok {
				m.SourceID = types.StringValue(id)
			}
		}
	}

	// Optional fields
	if name, ok := connection["name"].(string); connection["name"] != nil && ok {
		m.Name = types.StringValue(name)
	} else {
		m.Name = types.StringNull()
	}

	if description, ok := connection["description"].(string); connection["description"] != nil && ok {
		m.Description = types.StringValue(description)
	} else {
		m.Description = types.StringNull()
	}

	if disabledAt, ok := connection["disabled_at"].(string); connection["disabled_at"] != nil && ok {
		m.DisabledAt = types.StringValue(disabledAt)
	} else {
		m.DisabledAt = types.StringNull()
	}

	if pausedAt, ok := connection["paused_at"].(string); connection["paused_at"] != nil && ok {
		m.PausedAt = types.StringValue(pausedAt)
	} else {
		m.PausedAt = types.StringNull()
	}

	// Handle rules
	if rules, ok := connection["rules"].([]interface{}); ok && len(rules) > 0 {
		m.Rules = rulesFromAPI(rules)
	}
	// Keep rules as nil if not present or empty to maintain consistency with Terraform state

	return diags
}

func (m *connectionResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest(ctx, "GET", fmt.Sprintf("/%s/connections/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error reading connection", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading connection", string(body))
		} else {
			diags.AddError("Error reading connection", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading connection", err.Error())
		return diags
	}

	var connection map[string]interface{}
	err = json.Unmarshal(body, &connection)
	if err != nil {
		diags.AddError("Error reading connection", err.Error())
		return diags
	}

	return m.Refresh(connection)
}

func (m *connectionResourceModel) Create(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, err := m.toCreatePayload(ctx)
	if err != nil {
		diags.AddError("Error creating connection", err.Error())
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error creating connection", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "POST", fmt.Sprintf("/%s/connections", apiVersion), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error creating connection", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error creating connection", string(body))
		} else {
			diags.AddError("Error creating connection", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error creating connection", err.Error())
		return diags
	}

	var connection map[string]interface{}
	err = json.Unmarshal(body, &connection)
	if err != nil {
		diags.AddError("Error creating connection", err.Error())
		return diags
	}

	return m.Refresh(connection)
}

func (m *connectionResourceModel) Update(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, err := m.toUpdatePayload(ctx)
	if err != nil {
		diags.AddError("Error updating connection", err.Error())
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating connection", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "PUT", fmt.Sprintf("/%s/connections/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error updating connection", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error updating connection", string(body))
		} else {
			diags.AddError("Error updating connection", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error updating connection", err.Error())
		return diags
	}

	var connection map[string]interface{}
	err = json.Unmarshal(body, &connection)
	if err != nil {
		diags.AddError("Error updating connection", err.Error())
		return diags
	}

	return m.Refresh(connection)
}

func (m *connectionResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest(ctx, "DELETE", fmt.Sprintf("/%s/connections/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error deleting connection", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error deleting connection", string(body))
		} else {
			diags.AddError("Error deleting connection", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	return nil
}

// toCreatePayload converts the model to API create payload.
func (m *connectionResourceModel) toCreatePayload(ctx context.Context) (map[string]interface{}, error) {
	payload := map[string]interface{}{}

	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		payload["name"] = m.Name.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		payload["description"] = m.Description.ValueString()
	}
	if !m.DestinationID.IsNull() && !m.DestinationID.IsUnknown() {
		payload["destination_id"] = m.DestinationID.ValueString()
	}
	if !m.SourceID.IsNull() && !m.SourceID.IsUnknown() {
		payload["source_id"] = m.SourceID.ValueString()
	}

	// Convert rules - only include if not nil
	if m.Rules != nil {
		apiRules, err := rulesToAPI(ctx, m.Rules)
		if err != nil {
			return nil, err
		}
		payload["rules"] = apiRules
	}

	return payload, nil
}

// toUpdatePayload converts the model to API update payload.
func (m *connectionResourceModel) toUpdatePayload(ctx context.Context) (map[string]interface{}, error) {
	payload := map[string]interface{}{}

	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		payload["name"] = m.Name.ValueString()
	} else {
		payload["name"] = nil
	}

	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		payload["description"] = m.Description.ValueString()
	} else {
		payload["description"] = nil
	}

	// Convert rules - only include if not nil
	if m.Rules != nil {
		if len(m.Rules) > 0 {
			apiRules, err := rulesToAPI(ctx, m.Rules)
			if err != nil {
				return nil, err
			}
			payload["rules"] = apiRules
		} else {
			payload["rules"] = []interface{}{}
		}
	}

	return payload, nil
}

// rulesToAPI converts rules to API format.
func rulesToAPI(ctx context.Context, rules []rule) ([]interface{}, error) {
	result := []interface{}{}

	for _, ruleItem := range rules {
		if ruleItem.DeduplicateRule != nil {
			rule := map[string]interface{}{
				"type":   "deduplicate",
				"window": ruleItem.DeduplicateRule.Window.ValueInt64(),
			}

			// Check for mutual exclusivity - defensive programming in case schema validation changes
			includeFieldsSet := !ruleItem.DeduplicateRule.IncludeFields.IsNull() && !ruleItem.DeduplicateRule.IncludeFields.IsUnknown()
			excludeFieldsSet := !ruleItem.DeduplicateRule.ExcludeFields.IsNull() && !ruleItem.DeduplicateRule.ExcludeFields.IsUnknown()

			if includeFieldsSet && excludeFieldsSet {
				// This should be caught by schema validation, but we're being defensive
				return nil, fmt.Errorf("deduplicate rule cannot have both include_fields and exclude_fields set")
			}

			// Add include_fields if present
			if includeFieldsSet {
				includeFields := []string{}
				ruleItem.DeduplicateRule.IncludeFields.ElementsAs(ctx, &includeFields, false)
				if len(includeFields) > 0 {
					rule["include_fields"] = includeFields
				}
			}

			// Add exclude_fields if present
			if excludeFieldsSet {
				excludeFields := []string{}
				ruleItem.DeduplicateRule.ExcludeFields.ElementsAs(ctx, &excludeFields, false)
				if len(excludeFields) > 0 {
					rule["exclude_fields"] = excludeFields
				}
			}

			result = append(result, rule)
		}

		if ruleItem.DelayRule != nil {
			rule := map[string]interface{}{
				"type":  "delay",
				"delay": ruleItem.DelayRule.Delay.ValueInt64(),
			}
			result = append(result, rule)
		}

		if ruleItem.FilterRule != nil {
			rule := map[string]interface{}{
				"type": "filter",
			}

			if ruleItem.FilterRule.Body != nil {
				rule["body"] = filterRulePropertyToAPI(ruleItem.FilterRule.Body)
			}
			if ruleItem.FilterRule.Headers != nil {
				rule["headers"] = filterRulePropertyToAPI(ruleItem.FilterRule.Headers)
			}
			if ruleItem.FilterRule.Path != nil {
				rule["path"] = filterRulePropertyToAPI(ruleItem.FilterRule.Path)
			}
			if ruleItem.FilterRule.Query != nil {
				rule["query"] = filterRulePropertyToAPI(ruleItem.FilterRule.Query)
			}

			result = append(result, rule)
		}

		if ruleItem.RetryRule != nil {
			rule := map[string]interface{}{
				"type":     "retry",
				"strategy": ruleItem.RetryRule.Strategy.ValueString(),
			}

			if !ruleItem.RetryRule.Count.IsNull() && !ruleItem.RetryRule.Count.IsUnknown() {
				rule["count"] = ruleItem.RetryRule.Count.ValueInt64()
			}
			if !ruleItem.RetryRule.Interval.IsNull() && !ruleItem.RetryRule.Interval.IsUnknown() {
				rule["interval"] = ruleItem.RetryRule.Interval.ValueInt64()
			}

			responseStatusCodesFieldsSet := !ruleItem.RetryRule.ResponseStatusCodes.IsNull() && !ruleItem.RetryRule.ResponseStatusCodes.IsUnknown()
			if responseStatusCodesFieldsSet {
				responseStatusCodes := []string{}
				ruleItem.RetryRule.ResponseStatusCodes.ElementsAs(ctx, &responseStatusCodes, false)
				if len(responseStatusCodes) > 0 {
					rule["response_status_codes"] = responseStatusCodes
				}
			}

			result = append(result, rule)
		}

		if ruleItem.TransformRule != nil {
			rule := map[string]interface{}{
				"type":              "transform",
				"transformation_id": ruleItem.TransformRule.TransformationID.ValueString(),
			}
			result = append(result, rule)
		}
	}

	return result, nil
}

// filterRulePropertyToAPI converts filter rule property to API format.
func filterRulePropertyToAPI(property *filterRuleProperty) interface{} {
	if property == nil {
		return nil
	}

	if !property.Boolean.IsNull() && !property.Boolean.IsUnknown() {
		return property.Boolean.ValueBool()
	}

	if !property.JSON.IsNull() && !property.JSON.IsUnknown() {
		// jsontypes.Normalized has Unmarshal method that returns the parsed data
		var jsonData interface{}
		if diags := property.JSON.Unmarshal(&jsonData); !diags.HasError() {
			return jsonData
		}
	}

	if !property.Number.IsNull() && !property.Number.IsUnknown() {
		if float, _ := property.Number.ValueBigFloat().Float64(); true {
			return float
		}
	}

	if !property.String.IsNull() && !property.String.IsUnknown() {
		return property.String.ValueString()
	}

	return nil
}

// rulesFromAPI converts API rules to model format.
func rulesFromAPI(rules []interface{}) []rule {
	var result []rule

	for _, r := range rules {
		ruleMap, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		ruleType, _ := ruleMap["type"].(string)

		switch ruleType {
		case "deduplicate":
			deduplicateRule := &deduplicateRule{}

			if window, ok := ruleMap["window"].(float64); ok {
				deduplicateRule.Window = types.Int64Value(int64(window))
			}

			// Parse include_fields if present
			if includeFields, ok := ruleMap["include_fields"].([]interface{}); ok {
				fields := []types.String{}
				for _, field := range includeFields {
					if fieldStr, ok := field.(string); ok {
						fields = append(fields, types.StringValue(fieldStr))
					}
				}
				if len(fields) > 0 {
					deduplicateRule.IncludeFields, _ = types.ListValueFrom(context.Background(), types.StringType, fields)
				} else {
					deduplicateRule.IncludeFields = types.ListNull(types.StringType)
				}
			} else {
				deduplicateRule.IncludeFields = types.ListNull(types.StringType)
			}

			// Parse exclude_fields if present
			if excludeFields, ok := ruleMap["exclude_fields"].([]interface{}); ok {
				fields := []types.String{}
				for _, field := range excludeFields {
					if fieldStr, ok := field.(string); ok {
						fields = append(fields, types.StringValue(fieldStr))
					}
				}
				if len(fields) > 0 {
					deduplicateRule.ExcludeFields, _ = types.ListValueFrom(context.Background(), types.StringType, fields)
				} else {
					deduplicateRule.ExcludeFields = types.ListNull(types.StringType)
				}
			} else {
				deduplicateRule.ExcludeFields = types.ListNull(types.StringType)
			}

			result = append(result, rule{DeduplicateRule: deduplicateRule})

		case "delay":
			if delay, ok := ruleMap["delay"].(float64); ok {
				result = append(result, rule{
					DelayRule: &delayRule{
						Delay: types.Int64Value(int64(delay)),
					},
				})
			}

		case "filter":
			fr := &filterRule{}

			if body := ruleMap["body"]; body != nil {
				fr.Body = filterRulePropertyFromAPI(body)
			}
			if headers := ruleMap["headers"]; headers != nil {
				fr.Headers = filterRulePropertyFromAPI(headers)
			}
			if path := ruleMap["path"]; path != nil {
				fr.Path = filterRulePropertyFromAPI(path)
			}
			if query := ruleMap["query"]; query != nil {
				fr.Query = filterRulePropertyFromAPI(query)
			}

			result = append(result, rule{FilterRule: fr})

		case "retry":
			retryRule := &retryRule{}

			if strategy, ok := ruleMap["strategy"].(string); ok {
				retryRule.Strategy = types.StringValue(strategy)
			}
			if count, ok := ruleMap["count"].(float64); ok {
				retryRule.Count = types.Int64Value(int64(count))
			} else {
				retryRule.Count = types.Int64Null()
			}
			if interval, ok := ruleMap["interval"].(float64); ok {
				retryRule.Interval = types.Int64Value(int64(interval))
			} else {
				retryRule.Interval = types.Int64Null()
			}
			if responseStatusCodes, ok := ruleMap["response_status_codes"].([]any); ok {
				statusCodeExpressions := []types.String{}
				for _, expression := range responseStatusCodes {
					if expressionStr, ok := expression.(string); ok {
						statusCodeExpressions = append(statusCodeExpressions, types.StringValue(expressionStr))
					}
				}
				if len(statusCodeExpressions) > 0 {
					retryRule.ResponseStatusCodes, _ = types.ListValueFrom(context.Background(), types.StringType, statusCodeExpressions)
				} else {
					retryRule.ResponseStatusCodes = types.ListNull(types.StringType)
				}
			} else {
				retryRule.ResponseStatusCodes = types.ListNull(types.StringType)
			}

			result = append(result, rule{RetryRule: retryRule})

		case "transform":
			if transformationID, ok := ruleMap["transformation_id"].(string); ok {
				result = append(result, rule{
					TransformRule: &transformRule{
						TransformationID: types.StringValue(transformationID),
					},
				})
			}
		}
	}

	return result
}

// filterRulePropertyFromAPI converts API filter rule property to model format.
func filterRulePropertyFromAPI(property interface{}) *filterRuleProperty {
	if property == nil {
		return nil
	}

	result := &filterRuleProperty{}

	switch v := property.(type) {
	case bool:
		result.Boolean = types.BoolValue(v)
	case float64:
		result.Number = types.NumberValue(big.NewFloat(v))
	case string:
		result.String = types.StringValue(v)
	case map[string]interface{}, []interface{}:
		// jsontypes.Normalized will handle normalization automatically
		if jsonBytes, err := json.Marshal(v); err == nil {
			result.JSON = jsontypes.NewNormalizedValue(string(jsonBytes))
		}
	}

	return result
}
