package transformation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiVersion = "2025-07-01"

func (m *transformationResourceModel) Refresh(transformation map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Required fields
	if code, ok := transformation["code"].(string); ok {
		m.Code = types.StringValue(code)
	} else {
		diags.AddError("Error parsing code", "Expected string value")
		return diags
	}

	if createdAt, ok := transformation["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(createdAt)
	} else {
		diags.AddError("Error parsing created_at", "Expected string value")
		return diags
	}

	if id, ok := transformation["id"].(string); ok {
		m.ID = types.StringValue(id)
	} else {
		diags.AddError("Error parsing id", "Expected string value")
		return diags
	}

	if name, ok := transformation["name"].(string); ok {
		m.Name = types.StringValue(name)
	} else {
		diags.AddError("Error parsing name", "Expected string value")
		return diags
	}

	if teamID, ok := transformation["team_id"].(string); ok {
		m.TeamID = types.StringValue(teamID)
	} else {
		diags.AddError("Error parsing team_id", "Expected string value")
		return diags
	}

	if updatedAt, ok := transformation["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(updatedAt)
	} else {
		diags.AddError("Error parsing updated_at", "Expected string value")
		return diags
	}

	// Optional field - env
	if env, ok := transformation["env"].(map[string]interface{}); ok && len(env) > 0 {
		// Convert map[string]interface{} to map[string]string
		envMap := make(map[string]string)
		for k, v := range env {
			if strVal, ok := v.(string); ok {
				envMap[k] = strVal
			}
		}
		if len(envMap) > 0 {
			envBytes, err := json.Marshal(envMap)
			if err == nil {
				m.ENV = types.StringValue(string(envBytes))
			} else {
				m.ENV = types.StringNull()
			}
		} else {
			m.ENV = types.StringNull()
		}
	} else {
		m.ENV = types.StringNull()
	}

	return diags
}

func (m *transformationResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest("GET", fmt.Sprintf("/%s/transformations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error reading transformation", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading transformation", string(body))
		} else {
			diags.AddError("Error reading transformation", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading transformation", err.Error())
		return diags
	}

	var transformation map[string]interface{}
	err = json.Unmarshal(body, &transformation)
	if err != nil {
		diags.AddError("Error reading transformation", err.Error())
		return diags
	}

	return m.Refresh(transformation)
}

func (m *transformationResourceModel) Create(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload := m.toCreatePayload()

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error creating transformation", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("POST", fmt.Sprintf("/%s/transformations", apiVersion), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error creating transformation", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error creating transformation", string(body))
		} else {
			diags.AddError("Error creating transformation", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error creating transformation", err.Error())
		return diags
	}

	var transformation map[string]interface{}
	err = json.Unmarshal(body, &transformation)
	if err != nil {
		diags.AddError("Error creating transformation", err.Error())
		return diags
	}

	return m.Refresh(transformation)
}

func (m *transformationResourceModel) Update(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload := m.toUpdatePayload()

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating transformation", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("PUT", fmt.Sprintf("/%s/transformations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error updating transformation", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error updating transformation", string(body))
		} else {
			diags.AddError("Error updating transformation", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error updating transformation", err.Error())
		return diags
	}

	var transformation map[string]interface{}
	err = json.Unmarshal(body, &transformation)
	if err != nil {
		diags.AddError("Error updating transformation", err.Error())
		return diags
	}

	return m.Refresh(transformation)
}

func (m *transformationResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest("DELETE", fmt.Sprintf("/%s/transformations/%s", apiVersion, m.ID.ValueString()), &sdkclient.RequestOptions{})
	if err != nil {
		diags.AddError("Error deleting transformation", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error deleting transformation", string(body))
		} else {
			diags.AddError("Error deleting transformation", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	return nil
}

// toCreatePayload converts the model to API create payload.
func (m *transformationResourceModel) toCreatePayload() map[string]interface{} {
	payload := map[string]interface{}{
		"name": m.Name.ValueString(),
		"code": m.Code.ValueString(),
	}

	// Add env if present
	if !m.ENV.IsNull() && !m.ENV.IsUnknown() {
		env := m.getENV()
		if env != nil && len(env) > 0 {
			payload["env"] = env
		}
	}

	return payload
}

// toUpdatePayload converts the model to API update payload.
func (m *transformationResourceModel) toUpdatePayload() map[string]interface{} {
	payload := map[string]interface{}{}

	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		payload["name"] = m.Name.ValueString()
	} else {
		payload["name"] = nil
	}

	if !m.Code.IsNull() && !m.Code.IsUnknown() {
		payload["code"] = m.Code.ValueString()
	} else {
		payload["code"] = nil
	}

	// Handle env - API expects an object, not null
	if !m.ENV.IsNull() && !m.ENV.IsUnknown() {
		env := m.getENV()
		if env != nil && len(env) > 0 {
			payload["env"] = env
		} else {
			// Send empty object when clearing env
			payload["env"] = map[string]string{}
		}
	} else {
		// When env is null, send empty object to clear it
		payload["env"] = map[string]string{}
	}

	return payload
}

// getENV parses the JSON string into a map.
func (m *transformationResourceModel) getENV() map[string]string {
	envData := map[string]string{}
	if !m.ENV.IsUnknown() && !m.ENV.IsNull() {
		envBytes := []byte(m.ENV.ValueString())
		if err := json.Unmarshal(envBytes, &envData); err != nil {
			return nil
		}
	}
	return envData
}
