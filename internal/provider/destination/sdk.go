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
	"terraform-provider-hookdeck/internal/configjson"
	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiVersion = sdkclient.APIVersion

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
		m.ProjectID = types.StringValue(teamID)
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

// Retrieve reads the resource. With projectGoneOK a missing project counts as a
// missing resource. A resource's refresh sets it; a data source does not, so
// a wrong project_id keeps saying the project does not exist.
func (m *destinationResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client, projectGoneOK bool) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	response, err := client.RawClient.SendRequest(ctx, "GET", fmt.Sprintf("/%s/destinations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		QueryParams: url.Values{
			"include": []string{"config.auth"},
		},
	})
	if err != nil {
		if projectGoneOK && sdkclient.IsProjectGone(err) {
			return false, diags
		}
		diags.AddError("Error reading destination", err.Error())
		return false, diags
	}

	if sdkclient.IsGoneStatus(response.StatusCode) {
		return false, diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading destination", string(body))
		} else {
			diags.AddError("Error reading destination", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return false, diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading destination", err.Error())
		return false, diags
	}

	destination, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error reading destination", err.Error())
		return false, diags
	}

	diags.Append(m.Refresh(destination)...)
	if diags.HasError() {
		return false, diags
	}
	config, configDiags := shared.ConfigFromAPI(destination)
	diags.Append(configDiags...)
	m.Config = config
	return true, diags
}

// Create creates the destination from the configuration the user wrote.
// m.Config holds the planned value, which is unknown when the configuration
// has none, and is then taken from the response.
func (m *destinationResourceModel) Create(ctx context.Context, client *sdkclient.Client, userConfig jsontypes.Normalized) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toPayload(userConfig)
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

	destination, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error creating destination", err.Error())
		return diags
	}

	return m.refreshAfterWrite(destination)
}

// refreshAfterWrite reads the response of a create or update. The planned
// config is kept, because Terraform requires the applied value to equal the
// planned one, unless it is unknown.
func (m *destinationResourceModel) refreshAfterWrite(destination map[string]interface{}) diag.Diagnostics {
	diags := m.Refresh(destination)
	if diags.HasError() || !m.Config.IsUnknown() {
		return diags
	}
	config, configDiags := shared.ConfigFromAPI(destination)
	diags.Append(configDiags...)
	m.Config = config
	return diags
}

// Update updates the destination from the configuration the user wrote.
// managed holds the config key paths sent by earlier writes, so keys removed
// from the configuration are cleared.
func (m *destinationResourceModel) Update(ctx context.Context, client *sdkclient.Client, userConfig jsontypes.Normalized, managed configjson.Paths, priorState *destinationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toUpdatePayload(userConfig, managed, priorState)
	if diags.HasError() {
		return diags
	}
	// The API requires config with the type for types that must have one,
	// and keeps the stored type and config when neither is sent.
	if _, hasConfig := payload["config"]; !hasConfig && priorState != nil && m.Type.Equal(priorState.Type) {
		delete(payload, "type")
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

	destination, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error updating destination", err.Error())
		return diags
	}

	return m.refreshAfterWrite(destination)
}

func (m *destinationResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest(ctx, "DELETE", fmt.Sprintf("/%s/destinations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		if sdkclient.IsProjectGone(err) {
			return nil
		}
		diags.AddError("Error deleting destination", err.Error())
		return diags
	}

	if sdkclient.IsGoneStatus(response.StatusCode) {
		return nil
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

// toPayload builds the create payload and the base of the update payload from
// the config the user wrote, never m.Config. m.Config can hold everything
// Hookdeck returned.
func (m *destinationResourceModel) toPayload(userConfig jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{}
	payload["name"] = m.Name.ValueString()
	if userConfig.ValueString() != "" {
		payloadConfig, err := configjson.Parse(userConfig.ValueString())
		if err != nil {
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
// config keys, and delivery policy members, that the provider sent earlier
// (managed) and the configuration no longer has. The Hookdeck API merges
// config updates, so omitted keys would otherwise be retained on the
// destination. priorState tells whether a cleared key held an array or a
// boolean, which the API rejects as null.
func (m *destinationResourceModel) toUpdatePayload(userConfig jsontypes.Normalized, managed configjson.Paths, priorState *destinationResourceModel) (map[string]interface{}, diag.Diagnostics) {
	payload, diags := m.toPayload(userConfig)
	if diags.HasError() {
		return nil, diags
	}

	if len(managed) == 0 {
		return payload, diags
	}

	var priorPayloadConfig map[string]interface{}
	if priorState != nil && !priorState.Config.IsNull() && !priorState.Config.IsUnknown() && priorState.Config.ValueString() != "" {
		var err error
		priorPayloadConfig, err = configjson.Parse(priorState.Config.ValueString())
		if err != nil {
			diags.AddError("Error parsing prior destination config", err.Error())
			return nil, diags
		}
	}

	newPayloadConfig, ok := payload["config"].(map[string]interface{})
	if !ok {
		newPayloadConfig = map[string]interface{}{}
		payload["config"] = newPayloadConfig
	}

	// Policy members are independently merged by the API. Groups themselves
	// are replaced, including their overrides, so only this level needs clearing.
	if next, ok := newPayloadConfig["delivery_policy"].(map[string]interface{}); ok {
		for key := range managed["delivery_policy"] {
			if _, exists := next[key]; !exists {
				next[key] = nil
			}
		}
	}

	typeDefaults := nonNullableConfigDefaults[m.Type.ValueString()]
	for key := range managed {
		if _, exists := newPayloadConfig[key]; exists {
			continue
		}
		// Arrays and booleans are uniformly non-nullable in this API. Sending
		// null returns 422. If we have a documented default for the field,
		// send that to reset; otherwise leave it omitted (which falls back to
		// the merge behavior — the only safe choice without a known default).
		switch priorPayloadConfig[key].(type) {
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
