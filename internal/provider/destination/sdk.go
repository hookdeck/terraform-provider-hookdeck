package destination

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiVersion = "2025-07-01"

func (m *destinationResourceModel) Refresh(destination map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if createdAt, ok := destination["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(createdAt)
	} else {
		diags.AddError("Error parsing created_at", "Expected string value")
		return diags
	}

	if description, ok := destination["description"].(string); destination["description"] != nil && ok {
		m.Description = types.StringValue(description)
	} else {
		m.Description = types.StringNull()
	}

	if disabledAt, ok := destination["disabled_at"].(string); destination["disabled_at"] != nil && ok {
		m.DisabledAt = types.StringValue(disabledAt)
	} else {
		m.DisabledAt = types.StringNull()
	}

	if id, ok := destination["id"].(string); ok {
		m.ID = types.StringValue(id)
	} else {
		diags.AddError("Error parsing id", "Expected string value")
		return diags
	}

	if name, ok := destination["name"].(string); ok {
		m.Name = types.StringValue(name)
	} else {
		diags.AddError("Error parsing name", "Expected string value")
		return diags
	}

	if teamID, ok := destination["team_id"].(string); ok {
		m.TeamID = types.StringValue(teamID)
	} else {
		diags.AddError("Error parsing team_id", "Expected string value")
		return diags
	}

	if destinationType, ok := destination["type"].(string); destination["type"] != nil && ok {
		m.Type = types.StringValue(destinationType)
	} else {
		m.Type = types.StringNull()
	}

	if updatedAt, ok := destination["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(updatedAt)
	} else {
		diags.AddError("Error parsing updated_at", "Expected string value")
		return diags
	}

	return diags
}

func (m *destinationResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	response, err := client.RawClient.SendRequest(ctx, "GET", fmt.Sprintf("/%s/destinations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		QueryParams: url.Values{
			"include": []string{"config.auth"},
		},
	})
	if err != nil {
		diags.AddError("Error reading destination", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading destination", string(body))
		} else {
			diags.AddError("Error reading destination", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading destination", err.Error())
		return diags
	}

	var destination map[string]interface{}
	err = json.Unmarshal(body, &destination)
	if err != nil {
		diags.AddError("Error reading destination", err.Error())
		return diags
	}

	return m.Refresh(destination)
}

func (m *destinationResourceModel) Create(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toPayload()
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error creating destination", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "POST", fmt.Sprintf("/%s/destinations", apiVersion), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error creating destination", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error creating destination", string(body))
		} else {
			diags.AddError("Error creating destination", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error creating destination", err.Error())
		return diags
	}

	var destination map[string]interface{}
	err = json.Unmarshal(body, &destination)
	if err != nil {
		diags.AddError("Error creating destination", err.Error())
		return diags
	}

	return m.Refresh(destination)
}

func (m *destinationResourceModel) Update(ctx context.Context, client *sdkclient.Client, priorState *destinationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toUpdatePayload(priorState)
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating destination", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "PUT", fmt.Sprintf("/%s/destinations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error updating destination", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error updating destination", string(body))
		} else {
			diags.AddError("Error updating destination", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error updating destination", err.Error())
		return diags
	}

	var destination map[string]interface{}
	err = json.Unmarshal(body, &destination)
	if err != nil {
		diags.AddError("Error updating destination", err.Error())
		return diags
	}

	return m.Refresh(destination)
}

func (m *destinationResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest(ctx, "DELETE", fmt.Sprintf("/%s/destinations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error deleting destination", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error deleting destination", string(body))
		} else {
			diags.AddError("Error deleting destination", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	return nil
}

func (m *destinationResourceModel) toPayload() (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{}
	payload["name"] = m.Name.ValueString()
	if m.Config.ValueString() != "" {
		var payloadConfig map[string]interface{}
		err := json.Unmarshal([]byte(m.Config.ValueString()), &payloadConfig)
		if err != nil {
			var diags diag.Diagnostics
			diags.AddError("Error creating destination", err.Error())
			return nil, diags
		}
		payload["config"] = payloadConfig
	}
	if m.Description.ValueString() != "" {
		payload["description"] = m.Description.ValueString()
	} else {
		payload["description"] = (*string)(nil)
	}
	if m.Type.ValueString() != "" {
		payload["type"] = m.Type.ValueString()
	}

	return payload, diags
}

// nonNullableConfigDefaults holds the documented default values for
// destination config fields that the Hookdeck API marks as non-nullable.
// For nullable fields, sending null clears them — but the API rejects null
// for non-nullable fields (e.g. boolean `path_forwarding_disabled`). To still
// honor a user removing such a field from their Terraform config, send the
// documented default explicitly so the field is reset rather than retained.
//
// Keep in sync with the OpenAPI spec when new destination types or
// non-nullable fields are introduced.
var nonNullableConfigDefaults = map[string]map[string]interface{}{
	"HTTP":     {"path_forwarding_disabled": false},
	"CLI":      {"path_forwarding_disabled": false},
	"MOCK_API": {"path_forwarding_disabled": false},
}

// toUpdatePayload builds the update payload, explicitly clearing top-level
// config keys that existed in the prior state but were removed in the plan.
// The Hookdeck API merges config updates, so omitted keys would otherwise be
// retained on the destination.
func (m *destinationResourceModel) toUpdatePayload(priorState *destinationResourceModel) (map[string]interface{}, diag.Diagnostics) {
	payload, diags := m.toPayload()
	if diags.HasError() {
		return nil, diags
	}

	if priorState == nil || priorState.Config.ValueString() == "" {
		return payload, diags
	}

	var priorPayloadConfig map[string]interface{}
	if err := json.Unmarshal([]byte(priorState.Config.ValueString()), &priorPayloadConfig); err != nil {
		diags.AddError("Error parsing prior destination config", err.Error())
		return nil, diags
	}

	newPayloadConfig, ok := payload["config"].(map[string]interface{})
	if !ok {
		newPayloadConfig = map[string]interface{}{}
		payload["config"] = newPayloadConfig
	}

	typeDefaults := nonNullableConfigDefaults[m.Type.ValueString()]
	for key, priorValue := range priorPayloadConfig {
		if _, exists := newPayloadConfig[key]; exists {
			continue
		}
		// Arrays and booleans are uniformly non-nullable in this API. Sending
		// null returns 422. If we have a documented default for the field,
		// send that to reset; otherwise leave it omitted (which falls back to
		// the merge behavior — the only safe choice without a known default).
		switch priorValue.(type) {
		case []interface{}, bool:
			if def, ok := typeDefaults[key]; ok {
				newPayloadConfig[key] = def
			}
			continue
		}
		newPayloadConfig[key] = nil
	}

	return payload, diags
}
