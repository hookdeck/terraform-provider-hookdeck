package source

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

func (m *sourceResourceModel) Refresh(source map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if createdAt, ok := source["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(createdAt)
	} else {
		diags.AddError("Error parsing created_at", "Expected string value")
		return diags
	}

	if description, ok := source["description"].(string); source["description"] != nil && ok {
		m.Description = types.StringValue(description)
	} else {
		m.Description = types.StringNull()
	}

	if disabledAt, ok := source["disabled_at"].(string); source["disabled_at"] != nil && ok {
		m.DisabledAt = types.StringValue(disabledAt)
	} else {
		m.DisabledAt = types.StringNull()
	}

	if id, ok := source["id"].(string); ok {
		m.ID = types.StringValue(id)
	} else {
		diags.AddError("Error parsing id", "Expected string value")
		return diags
	}

	if name, ok := source["name"].(string); ok {
		m.Name = types.StringValue(name)
	} else {
		diags.AddError("Error parsing name", "Expected string value")
		return diags
	}

	if teamID, ok := source["team_id"].(string); ok {
		m.ProjectID = types.StringValue(teamID)
	} else {
		diags.AddError("Error parsing team_id", "Expected string value")
		return diags
	}

	if sourceType, ok := source["type"].(string); source["type"] != nil && ok {
		m.Type = types.StringValue(sourceType)
	} else {
		m.Type = types.StringNull()
	}

	if updatedAt, ok := source["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(updatedAt)
	} else {
		diags.AddError("Error parsing updated_at", "Expected string value")
		return diags
	}

	if url, ok := source["url"].(string); ok {
		m.URL = types.StringValue(url)
	} else {
		diags.AddError("Error parsing url", "Expected string value")
		return diags
	}

	return diags
}

func (m *sourceResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	response, err := client.RawClient.SendRequest(ctx, "GET", fmt.Sprintf("/%s/sources/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		QueryParams: url.Values{
			"include": []string{"config.auth"},
		},
	})
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return false, diags
	}

	if sdkclient.IsGoneStatus(response.StatusCode) {
		return false, diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading source", string(body))
		} else {
			diags.AddError("Error reading source", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return false, diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return false, diags
	}

	source, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return false, diags
	}

	diags.Append(m.Refresh(source)...)
	if diags.HasError() {
		return false, diags
	}
	config, configDiags := shared.ConfigFromAPI(source)
	diags.Append(configDiags...)
	m.Config = config
	return true, diags
}

// Create creates the source from the configuration the user wrote. m.Config
// holds the planned value, which is unknown when the configuration has none,
// and is then taken from the response.
func (m *sourceResourceModel) Create(ctx context.Context, client *sdkclient.Client, userConfig jsontypes.Normalized) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toPayload(userConfig)
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "POST", fmt.Sprintf("/%s/sources", apiVersion), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error creating source", string(body))
		} else {
			diags.AddError("Error creating source", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	source, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	return m.refreshAfterWrite(source)
}

// refreshAfterWrite reads the response of a create or update. The planned
// config is kept, because Terraform requires the applied value to equal the
// planned one, unless it is unknown.
func (m *sourceResourceModel) refreshAfterWrite(source map[string]interface{}) diag.Diagnostics {
	diags := m.Refresh(source)
	if diags.HasError() || !m.Config.IsUnknown() {
		return diags
	}
	config, configDiags := shared.ConfigFromAPI(source)
	diags.Append(configDiags...)
	m.Config = config
	return diags
}

// Update updates the source from the configuration the user wrote. managed
// holds the config key paths sent by earlier writes, so keys removed from the
// configuration are cleared.
func (m *sourceResourceModel) Update(ctx context.Context, client *sdkclient.Client, userConfig jsontypes.Normalized, managed configjson.Paths, priorState *sourceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toUpdatePayload(userConfig, managed, priorState)
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest(ctx, "PUT", fmt.Sprintf("/%s/sources/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error updating source", string(body))
		} else {
			diags.AddError("Error updating source", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	source, err := shared.DecodeObject(body)
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	return m.refreshAfterWrite(source)
}

func (m *sourceResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest(ctx, "DELETE", fmt.Sprintf("/%s/sources/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error deleting source", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error deleting source", string(body))
		} else {
			diags.AddError("Error deleting source", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	return nil
}

// toPayload builds the create payload and the base of the update payload from
// the config the user wrote, never m.Config. m.Config can hold everything
// Hookdeck returned, including auth, and sending auth back would replace the
// authentication hookdeck_gateway_source_auth manages.
func (m *sourceResourceModel) toPayload(userConfig jsontypes.Normalized) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{}
	payload["name"] = m.Name.ValueString()
	if userConfig.ValueString() != "" {
		payloadConfig, err := configjson.Parse(userConfig.ValueString())
		if err != nil {
			diags.AddError("Error creating source", err.Error())
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

// nonNullableConfigDefaults holds the documented default values for source
// config fields that the Hookdeck API marks as non-nullable. For nullable
// fields, sending null clears them — but the API rejects null for
// non-nullable fields (e.g. array `allowed_http_methods`). To still honor a
// user removing such a field from their Terraform config, send the
// documented default explicitly so the field is reset rather than retained.
//
// Keep in sync with the OpenAPI spec when new source types or non-nullable
// fields are introduced.
var nonNullableConfigDefaults = map[string]map[string]interface{}{
	"HTTP": {
		"allowed_http_methods": []interface{}{"PUT", "POST", "PATCH", "DELETE"},
	},
}

// toUpdatePayload builds the update payload, explicitly clearing top-level
// config keys the provider sent earlier (managed) that the configuration no
// longer has. The Hookdeck API merges config updates, so omitted keys would
// otherwise be retained on the source. priorState supplies the type of the
// value being cleared.
func (m *sourceResourceModel) toUpdatePayload(userConfig jsontypes.Normalized, managed configjson.Paths, priorState *sourceResourceModel) (map[string]interface{}, diag.Diagnostics) {
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
			diags.AddError("Error parsing prior source config", err.Error())
			return nil, diags
		}
	}

	newPayloadConfig, ok := payload["config"].(map[string]interface{})
	if !ok {
		newPayloadConfig = map[string]interface{}{}
		payload["config"] = newPayloadConfig
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
