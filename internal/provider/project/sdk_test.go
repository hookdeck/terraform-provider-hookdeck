package project

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type response struct {
	status int
	body   string
}

// fakeAPI answers "METHOD path" from routes and records the calls.
type fakeAPI struct {
	routes map[string]response
	calls  []string
}

func (f *fakeAPI) SendRequest(_ context.Context, method, path string, _ *sdkclient.RequestOptions) (*http.Response, error) {
	call := method + " " + strings.TrimPrefix(path, "/"+sdkclient.APIVersion)
	f.calls = append(f.calls, call)
	r, ok := f.routes[call]
	if !ok {
		return nil, fmt.Errorf("unexpected request %s", call)
	}
	return &http.Response{StatusCode: r.status, Body: io.NopCloser(strings.NewReader(r.body))}, nil
}

func clientFor(kind projectscope.KeyKind, routes map[string]response) (sdkclient.Client, *fakeAPI) {
	api := &fakeAPI{routes: routes}
	return sdkclient.Client{RawClient: api, Scope: projectscope.Scope{KeyKind: kind}}, api
}

func projectJSON(id, name, apiType string) string {
	return fmt.Sprintf(`{"id":%q,"name":%q,"type":%q,"organization_id":"org_1","created_at":"2026-09-01T00:00:00Z","updated_at":"2026-09-01T00:00:00Z","headers_prefix":"x-set-elsewhere","notification_methods":["email"]}`, id, name, apiType)
}

func TestCreate_isOneRequest(t *testing.T) {
	client, api := clientFor(projectscope.KeyKindOrganization, map[string]response{
		"POST /projects": {200, projectJSON("tm_1", "prod", "event_gateway")},
	})
	m := projectResourceModel{Name: types.StringValue("prod")}

	if diags := m.create(t.Context(), gateway, client); diags.HasError() {
		t.Fatal(diags)
	}
	if m.ID.ValueString() != "tm_1" || m.OrganizationID.ValueString() != "org_1" || len(api.calls) != 1 {
		t.Errorf("model = %+v, calls = %v", m, api.calls)
	}
}

func TestCreate_responseWithoutID(t *testing.T) {
	client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{
		"POST /projects": {200, `{"name":"prod"}`},
	})
	m := projectResourceModel{Name: types.StringValue("prod")}

	if diags := m.create(t.Context(), gateway, client); !diags.HasError() {
		t.Error("expected an error")
	}
}

// Project settings are not managed here. An update that sent them, even
// as null, would overwrite what was set elsewhere.
func TestCreateAndUpdate_sendNoSettings(t *testing.T) {
	var sent string
	client := sdkclient.Client{RawClient: &captureAPI{body: projectJSON("tm_1", "prod", "event_gateway"), sent: &sent}}
	m := projectResourceModel{ID: types.StringValue("tm_1"), Name: types.StringValue("prod")}

	if diags := m.create(t.Context(), gateway, client); diags.HasError() {
		t.Fatal(diags)
	}
	if sent != `{"name":"prod","type":"event_gateway"}` {
		t.Errorf("create payload = %s", sent)
	}
	if diags := m.update(t.Context(), gateway, client); diags.HasError() {
		t.Fatal(diags)
	}
	if sent != `{"name":"prod"}` {
		t.Errorf("update payload = %s", sent)
	}
}

type captureAPI struct {
	body string
	sent *string
}

func (c *captureAPI) SendRequest(_ context.Context, _, _ string, opts *sdkclient.RequestOptions) (*http.Response, error) {
	if opts != nil && opts.Body != nil {
		data, _ := io.ReadAll(opts.Body)
		*c.sent = string(data)
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(c.body))}, nil
}

func TestRetrieve(t *testing.T) {
	cases := []struct {
		name      string
		response  response
		wantFound bool
		wantErr   string
	}{
		{"gateway project", response{200, projectJSON("tm_1", "prod", "event_gateway")}, true, ""},
		{"outpost project", response{200, projectJSON("tm_1", "prod", "outpost")}, true, "Unsupported project type"},
		{"console project", response{200, projectJSON("tm_1", "prod", "console")}, true, "Unsupported project type"},
		{"not found", response{404, `{"message":"Not Found"}`}, false, ""},
		{"no access", response{403, `{"code":"INSUFFICIENT_SCOPE"}`}, false, "Error reading project"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"GET /projects/tm_1": tc.response})
			m := projectResourceModel{ID: types.StringValue("tm_1")}

			found, diags := m.retrieve(t.Context(), gateway, client)

			if tc.wantErr == "" && diags.HasError() {
				t.Fatalf("diagnostics = %v", diags)
			}
			if tc.wantErr != "" && (!diags.HasError() || diags[0].Summary() != tc.wantErr) {
				t.Fatalf("diagnostics = %v, want %q", diags, tc.wantErr)
			}
			if found != tc.wantFound {
				t.Errorf("found = %v, want %v", found, tc.wantFound)
			}
			if tc.wantErr == "Unsupported project type" && !strings.Contains(diags[0].Detail(), "hookdeck_gateway_project") {
				t.Errorf("detail = %q", diags[0].Detail())
			}
		})
	}
}

func TestDelete_goneIsDeleted(t *testing.T) {
	for _, status := range []int{200, 404, 410} {
		client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"DELETE /projects/tm_1": {status, "{}"}})
		m := projectResourceModel{ID: types.StringValue("tm_1")}
		if diags := m.delete(t.Context(), client); diags.HasError() {
			t.Errorf("status %d: %v", status, diags)
		}
	}
	client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"DELETE /projects/tm_1": {500, "{}"}})
	m := projectResourceModel{ID: types.StringValue("tm_1")}
	if diags := m.delete(t.Context(), client); !diags.HasError() {
		t.Error("a failed delete must be an error")
	}
}

func TestFindByName(t *testing.T) {
	list := "[" + strings.Join([]string{
		projectJSON("tm_1", "prod", "event_gateway"),
		projectJSON("tm_2", "prod", "outpost"),
		projectJSON("tm_3", "staging", "event_gateway"),
		projectJSON("tm_4", "staging", "event_gateway"),
		projectJSON("tm_5", "deliveries", "outpost"),
	}, ",") + "]"

	cases := []struct {
		name, lookup, wantID, wantErr string
	}{
		{"a project of another type with the same name is ignored", "prod", "tm_1", ""},
		{"two gateway projects with the same name", "staging", "", "Ambiguous project name"},
		{"only a project of another type has the name", "deliveries", "", "Project not found"},
		{"no such name", "nope", "", "Project not found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := clientFor(projectscope.KeyKindOrganization, map[string]response{"GET /projects": {200, list}})
			project, diags := findByName(t.Context(), gateway, client, tc.lookup)
			if tc.wantErr != "" {
				if !diags.HasError() || diags[0].Summary() != tc.wantErr {
					t.Fatalf("diagnostics = %v, want %q", diags, tc.wantErr)
				}
				return
			}
			if diags.HasError() || project["id"] != tc.wantID {
				t.Fatalf("project = %v, diagnostics = %v", project["id"], diags)
			}
		})
	}
}

// With a project API key the resource refuses to plan or read: the API
// answers 404 for any project but the key's own, which would otherwise
// read as "deleted".
func TestResource_projectKeyIsRefused(t *testing.T) {
	client, api := clientFor(projectscope.KeyKindProject, nil)
	r := &projectResource{kind: gateway, client: client}

	planResp := &resource.ModifyPlanResponse{}
	r.ModifyPlan(t.Context(), resource.ModifyPlanRequest{}, planResp)
	if !planResp.Diagnostics.HasError() || planResp.Diagnostics[0].Summary() != "Organization API key required" {
		t.Errorf("plan diagnostics = %v", planResp.Diagnostics)
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(t.Context(), resource.SchemaRequest{}, schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(schemaResp.Schema.Type().TerraformType(t.Context()), nil)}
	readResp := &resource.ReadResponse{State: state}
	r.Read(t.Context(), resource.ReadRequest{State: state}, readResp)
	if !readResp.Diagnostics.HasError() || readResp.Diagnostics[0].Summary() != "Organization API key required" {
		t.Errorf("read diagnostics = %v", readResp.Diagnostics)
	}
	if len(api.calls) != 0 {
		t.Errorf("calls = %v", api.calls)
	}

	orgClient, _ := clientFor(projectscope.KeyKindOrganization, nil)
	planResp = &resource.ModifyPlanResponse{}
	(&projectResource{kind: gateway, client: orgClient}).ModifyPlan(t.Context(), resource.ModifyPlanRequest{}, planResp)
	if planResp.Diagnostics.HasError() {
		t.Errorf("organization key: %v", planResp.Diagnostics)
	}
}
