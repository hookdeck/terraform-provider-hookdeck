package project

import (
	"context"
	"fmt"

	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *projectResourceModel) refresh(project map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	str := func(key string, required bool) types.String {
		if v, ok := project[key].(string); ok {
			return types.StringValue(v)
		}
		if required {
			diags.AddError("Error parsing project", fmt.Sprintf("Expected string for %q", key))
		}
		return types.StringNull()
	}

	m.ID = str("id", true)
	m.Name = str("name", true)
	m.OrganizationID = str("organization_id", true)
	m.Type = str("type", true)
	m.CreatedAt = str("created_at", true)
	m.UpdatedAt = str("updated_at", true)
	m.HeadersPrefix = str("headers_prefix", false)

	if v, ok := project["max_events_per_second"].(float64); ok {
		m.MaxEventsPerSecond = types.Int64Value(int64(v))
	} else {
		m.MaxEventsPerSecond = types.Int64Null()
	}

	methods := []attr.Value{}
	if raw, ok := project["notification_methods"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				methods = append(methods, types.StringValue(s))
			}
		}
	}
	list, listDiags := types.ListValue(types.StringType, methods)
	diags.Append(listDiags...)
	m.NotificationMethods = list

	return diags
}

func (m *projectResourceModel) retrieve(ctx context.Context, client sdkclient.Client) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var project map[string]interface{}
	err := shared.Request(ctx, client, "GET", "/projects/"+m.ID.ValueString(), nil, nil, &project)
	if shared.IsNotFound(err) {
		return false, diags
	}
	if err != nil {
		diags.AddError("Error reading project", err.Error())
		return false, diags
	}
	return true, m.refresh(project)
}

func (m *projectResourceModel) create(ctx context.Context, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	var project map[string]interface{}
	err := shared.Request(ctx, client, "POST", "/projects", map[string]interface{}{
		"name": m.Name.ValueString(),
		"type": projectType,
	}, nil, &project)
	if err != nil {
		diags.AddError("Error creating project", err.Error())
		return diags
	}
	id, _ := project["id"].(string)
	m.ID = types.StringValue(id)

	// Settings live on the update endpoint.
	if !m.HeadersPrefix.IsNull() || !m.NotificationMethods.IsNull() && !m.NotificationMethods.IsUnknown() {
		return m.update(ctx, client)
	}
	return m.refresh(project)
}

func (m *projectResourceModel) update(ctx context.Context, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	payload := map[string]interface{}{
		"name":           m.Name.ValueString(),
		"headers_prefix": nil,
	}
	if !m.HeadersPrefix.IsNull() {
		payload["headers_prefix"] = m.HeadersPrefix.ValueString()
	}
	if !m.NotificationMethods.IsNull() && !m.NotificationMethods.IsUnknown() {
		var methods []string
		diags.Append(m.NotificationMethods.ElementsAs(ctx, &methods, false)...)
		if diags.HasError() {
			return diags
		}
		payload["notification_methods"] = methods
	}
	var project map[string]interface{}
	if err := shared.Request(ctx, client, "PUT", "/projects/"+m.ID.ValueString(), payload, nil, &project); err != nil {
		diags.AddError("Error updating project", err.Error())
		return diags
	}
	diags.Append(m.refresh(project)...)
	return diags
}

func (m *projectResourceModel) delete(ctx context.Context, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	err := shared.Request(ctx, client, "DELETE", "/projects/"+m.ID.ValueString(), nil, nil, nil)
	if err != nil && !shared.IsNotFound(err) {
		diags.AddError("Error deleting project", err.Error())
	}
	return diags
}

// findByName lists the projects the key can see and returns the one named
// name. Exactly one match is required.
func findByName(ctx context.Context, client sdkclient.Client, name string) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var projects []map[string]interface{}
	if err := shared.Request(ctx, client, "GET", "/projects", nil, nil, &projects); err != nil {
		diags.AddError("Error listing projects", err.Error())
		return nil, diags
	}
	var matches []map[string]interface{}
	for _, p := range projects {
		if n, _ := p["name"].(string); n == name {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], diags
	case 0:
		diags.AddError("Project not found", fmt.Sprintf("No project named %q is visible to this API key.", name))
	default:
		diags.AddError("Ambiguous project name", fmt.Sprintf("%d projects are named %q; look the project up by id instead.", len(matches), name))
	}
	return nil, diags
}
