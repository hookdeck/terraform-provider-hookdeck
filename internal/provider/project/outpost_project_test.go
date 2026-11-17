package project

import (
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRetrieve_outpostKindRefusesGatewayProject(t *testing.T) {
	cases := []struct {
		name, apiType, wantErr string
	}{
		{"outpost project", "outpost", ""},
		{"gateway project", "event_gateway", "Unsupported project type"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"GET /projects/tm_1": {200, projectJSON("tm_1", "prod", tc.apiType)}})
			m := projectResourceModel{ID: types.StringValue("tm_1")}

			found, diags := m.retrieve(t.Context(), outpost, client)

			if tc.wantErr == "" {
				if diags.HasError() || !found {
					t.Fatalf("found = %v, diagnostics = %v", found, diags)
				}
				return
			}
			if !diags.HasError() || diags[0].Summary() != tc.wantErr || !strings.Contains(diags[0].Detail(), "hookdeck_outpost_project") {
				t.Fatalf("diagnostics = %v", diags)
			}
		})
	}
}

func TestFindByName_outpostKindIgnoresGatewayProjects(t *testing.T) {
	list := "[" + strings.Join([]string{
		projectJSON("tm_1", "prod", "event_gateway"),
		projectJSON("tm_2", "prod", "outpost"),
		projectJSON("tm_3", "events", "event_gateway"),
	}, ",") + "]"
	client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"GET /projects": {200, list}})

	project, diags := findByName(t.Context(), outpost, client, "prod")
	if diags.HasError() || project["id"] != "tm_2" {
		t.Fatalf("project = %v, diagnostics = %v", project["id"], diags)
	}
	if _, diags := findByName(t.Context(), outpost, client, "events"); !diags.HasError() || diags[0].Summary() != "Project not found" {
		t.Errorf("diagnostics = %v", diags)
	}
}
