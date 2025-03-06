package sourceauth

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

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiVersion = "2025-01-01"

func (m *sourceAuthResourceModel) Refresh(source map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	config, ok := source["config"].(map[string]interface{})
	if !ok {
		diags.AddError("Error parsing config", "Expected map[string]interface{} value")
		return diags
	}

	if authType, ok := config["auth_type"].(string); config["auth_type"] != nil && ok {
		m.AuthType = types.StringValue(authType)
	} else {
		m.AuthType = types.StringNull()
	}

	if config["auth"] != nil {
		auth, err := json.Marshal(config["auth"])
		if err != nil {
			diags.AddError("Error marshalling auth", err.Error())
			return diags
		}
		m.Auth = jsontypes.NewNormalizedValue(string(auth))
	} else {
		m.Auth = jsontypes.NewNormalizedValue("{}")
	}

	return diags
}

func (m *sourceAuthResourceModel) doRetrieve(_ context.Context, client *sdkclient.Client) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	response, err := client.RawClient.SendRequest("GET", fmt.Sprintf("/%s/sources/%s", apiVersion, m.SourceID.ValueString()), &sdkclient.RequestOptions{
		QueryParams: url.Values{
			"include": []string{"config.auth"},
		},
	})
	if err != nil {
		diags.AddError("Error reading source auth", err.Error())
		return nil, diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error reading source auth", string(body))
		} else {
			diags.AddError("Error reading source auth", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return nil, diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error reading source auth", err.Error())
		return nil, diags
	}

	var source map[string]interface{}
	err = json.Unmarshal(body, &source)
	if err != nil {
		diags.AddError("Error reading source auth", err.Error())
		return nil, diags
	}

	return source, diags
}

func (m *sourceAuthResourceModel) Retrieve(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	source, diags := m.doRetrieve(ctx, client)
	if diags.HasError() {
		return diags
	}

	return m.Refresh(source)
}

func (m *sourceAuthResourceModel) Create(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	return m.Update(ctx, client)
}

func (m *sourceAuthResourceModel) Update(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	initSource, diags := m.doRetrieve(ctx, client)
	if diags.HasError() {
		return diags
	}

	payload, diags := m.toPayload(initSource)
	if diags.HasError() {
		return diags
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error updating source auth", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("PUT", fmt.Sprintf("/%s/sources/%s", apiVersion, m.SourceID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error updating source auth", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error updating source auth", string(body))
		} else {
			diags.AddError("Error updating source auth", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		diags.AddError("Error updating source auth", err.Error())
		return diags
	}

	var source map[string]interface{}
	err = json.Unmarshal(body, &source)
	if err != nil {
		diags.AddError("Error updating source auth", err.Error())
		return diags
	}

	return m.Refresh(source)
}

func (m *sourceAuthResourceModel) Delete(ctx context.Context, client *sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	initSource, diags := m.doRetrieve(ctx, client)
	if diags.HasError() {
		return diags
	}

	config, ok := initSource["config"].(map[string]interface{})
	if !ok {
		diags.AddError("Error parsing config", "Expected map[string]interface{} value")
		return diags
	}

	config["auth"] = (*map[string]interface{})(nil)
	config["auth_type"] = (*string)(nil)
	payload := map[string]interface{}{
		"type":   initSource["type"],
		"config": config,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Error deleting source auth", err.Error())
		return diags
	}

	response, err := client.RawClient.SendRequest("PUT", fmt.Sprintf("/sources/%s", m.SourceID.ValueString()), &sdkclient.RequestOptions{
		Body: bytes.NewReader(jsonData),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	})
	if err != nil {
		diags.AddError("Error deleting source auth", err.Error())
		return diags
	}

	if response.StatusCode > 299 {
		if body, err := io.ReadAll(response.Body); err == nil {
			diags.AddError("Error deleting source auth", string(body))
		} else {
			diags.AddError("Error deleting source auth", "Status code: "+strconv.Itoa(response.StatusCode))
		}
		return diags
	}

	return nil
}

func (m *sourceAuthResourceModel) toPayload(source map[string]interface{}) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	config, ok := source["config"].(map[string]interface{})
	if !ok {
		diags.AddError("Error parsing config", "Expected map[string]interface{} value")
		return nil, diags
	}

	delete(config, "auth")
	delete(config, "auth_type")

	if m.AuthType.ValueString() != "" {
		config["auth_type"] = m.AuthType.ValueString()
	}
	if m.Auth.ValueString() != "" {
		var auth map[string]interface{}
		err := json.Unmarshal([]byte(m.Auth.ValueString()), &auth)
		if err != nil {
			diags.AddError("Error unmarshalling source auth", err.Error())
			return nil, diags
		}
		config["auth"] = auth
	}

	payload := map[string]interface{}{
		"type":   source["type"],
		"config": config,
	}
	return payload, diags
}
