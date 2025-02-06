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
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		m.TeamID = types.StringValue(teamID)
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

func (m *sourceResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	response, err := client.RawClient.SendRequest("GET", fmt.Sprintf("/sources/%s", m.ID.ValueString()), &sdkclient.RequestOptions{
		QueryParams: url.Values{
			"include": []string{"config.auth"},
		},
	})
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading source", string(body))
		} else {
			diags.AddError("Error reading source", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return diags
	}

	var source map[string]interface{}
	err = json.Unmarshal(body, &source)
	if err != nil {
		diags.AddError("Error reading source", err.Error())
		return diags
	}

	return m.Refresh(source)
}

func (m *sourceResourceModel) Create(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toPayload()
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("POST", "/sources", &sdkclient.RequestOptions{
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

	var source map[string]interface{}
	err = json.Unmarshal(body, &source)
	if err != nil {
		diags.AddError("Error creating source", err.Error())
		return diags
	}

	return m.Refresh(source)
}

func (m *sourceResourceModel) Update(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	payload, diags := m.toPayload()
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("PUT", fmt.Sprintf("/sources/%s", m.ID.ValueString()), &sdkclient.RequestOptions{
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

	var source map[string]interface{}
	err = json.Unmarshal(body, &source)
	if err != nil {
		diags.AddError("Error updating source", err.Error())
		return diags
	}

	return m.Refresh(source)
}

func (m *sourceResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	response, err := client.RawClient.SendRequest("DELETE", fmt.Sprintf("/sources/%s", m.ID.ValueString()), &sdkclient.RequestOptions{})
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

func (m *sourceResourceModel) toPayload() (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := map[string]interface{}{}
	payload["name"] = m.Name.ValueString()
	if m.Config.ValueString() != "" {
		var payloadConfig map[string]interface{}
		err := json.Unmarshal([]byte(m.Config.ValueString()), &payloadConfig)
		if err != nil {
			var diags diag.Diagnostics
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
