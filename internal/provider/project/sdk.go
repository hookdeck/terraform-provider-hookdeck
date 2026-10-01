package project

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *projectResourceModel) refresh(k kind, project map[string]any) diag.Diagnostics {
	var diags diag.Diagnostics

	str := func(key string) types.String {
		if v, ok := project[key].(string); ok {
			return types.StringValue(v)
		}
		diags.AddError("Error parsing project", fmt.Sprintf("Expected string for %q", key))
		return types.StringNull()
	}

	id := str("id")
	apiType := str("type")
	if diags.HasError() {
		return diags
	}
	if apiType.ValueString() != k.apiType {
		diags.AddError("Unsupported project type", k.wrongType(id.ValueString(), apiType.ValueString()))
		return diags
	}

	m.ID = id
	m.Type = apiType
	m.Name = str("name")
	m.OrganizationID = str("organization_id")
	m.CreatedAt = str("created_at")
	m.UpdatedAt = str("updated_at")
	return diags
}

func (m *projectResourceModel) retrieve(ctx context.Context, k kind, client sdkclient.Client) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	var project map[string]any
	err := client.Do(ctx, http.MethodGet, "/projects/"+m.ID.ValueString(), nil, nil, &project)
	if sdkclient.IsNotFound(err) {
		return false, diags
	}
	if err != nil {
		diags.AddError("Error reading project", err.Error())
		return false, diags
	}
	return true, m.refresh(k, project)
}

func (m *projectResourceModel) create(ctx context.Context, k kind, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	var project map[string]any
	err := client.Do(ctx, http.MethodPost, "/projects", map[string]any{
		"name": m.Name.ValueString(),
		"type": k.apiType,
	}, nil, &project)
	if err != nil {
		diags.AddError("Error creating project", err.Error())
		return diags
	}
	return m.refresh(k, project)
}

// update sends the name only: the API merges by presence, so project
// settings made elsewhere are left as they are.
func (m *projectResourceModel) update(ctx context.Context, k kind, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	var project map[string]any
	err := client.Do(ctx, http.MethodPut, "/projects/"+m.ID.ValueString(), map[string]any{
		"name": m.Name.ValueString(),
	}, nil, &project)
	if err != nil {
		diags.AddError("Error updating project", err.Error())
		return diags
	}
	return m.refresh(k, project)
}

func (m *projectResourceModel) delete(ctx context.Context, client sdkclient.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	err := client.Do(ctx, http.MethodDelete, "/projects/"+m.ID.ValueString(), nil, nil, nil)
	if err != nil && !sdkclient.IsNotFound(err) {
		diags.AddError("Error deleting project", err.Error())
	}
	return diags
}

// findByName lists the projects of kind k the key can see and returns the
// one named name. Exactly one match is required.
func findByName(ctx context.Context, k kind, client sdkclient.Client, name string) (map[string]any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var projects []map[string]any
	if err := client.Do(ctx, http.MethodGet, "/projects", nil, nil, &projects); err != nil {
		diags.AddError("Error listing projects", err.Error())
		return nil, diags
	}
	var matches []map[string]any
	for _, p := range projects {
		if p["name"] == name && p["type"] == k.apiType {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], diags
	case 0:
		diags.AddError("Project not found", fmt.Sprintf("No %s project named %q is visible to this API key.", k.label, name))
	default:
		diags.AddError("Ambiguous project name", fmt.Sprintf("%d %s projects are named %q; look the project up by ID instead.", len(matches), k.label, name))
	}
	return nil, diags
}
