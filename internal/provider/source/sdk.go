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
	m.CreatedAt = types.StringValue(source["created_at"].(string))
	if source["description"] != nil {
		m.Description = types.StringValue(source["description"].(string))
	} else {
		m.Description = types.StringNull()
	}
	if source["disabled_at"] != nil {
		m.DisabledAt = types.StringValue(source["disabled_at"].(string))
	} else {
		m.DisabledAt = types.StringNull()
	}
	m.ID = types.StringValue(source["id"].(string))
	m.Name = types.StringValue(source["name"].(string))
	m.TeamID = types.StringValue(source["team_id"].(string))
	if source["type"] != nil {
		m.Type = types.StringValue(source["type"].(string))
	} else {
		m.Type = types.StringNull()
	}
	m.UpdatedAt = types.StringValue(source["updated_at"].(string))
	m.URL = types.StringValue(source["url"].(string))
	return nil
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
