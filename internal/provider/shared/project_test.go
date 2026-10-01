package shared

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// recordingClient answers 200 {} and records each request's project header
// and path. Paths listed in missing answer 404.
type recordingClient struct {
	projects []string
	paths    []string
	missing  map[string]bool
}

func (c *recordingClient) SendRequest(_ context.Context, _, path string, opts *sdkclient.RequestOptions) (*http.Response, error) {
	project := ""
	if opts != nil {
		project = opts.Headers.Get(sdkclient.ProjectHeader)
	}
	c.projects = append(c.projects, project)
	c.paths = append(c.paths, path)
	status, body := http.StatusOK, "{}"
	if c.missing[path] {
		status, body = http.StatusNotFound, `{"message":"Not Found"}`
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
}

var (
	projectKey        = projectscope.Scope{KeyKind: projectscope.KeyKindProject, ProjectID: "tm_key"}
	projectKeyUnknown = projectscope.Scope{KeyKind: projectscope.KeyKindProject}
	orgSingle         = projectscope.Scope{KeyKind: projectscope.KeyKindOrganization, ProjectID: "tm_prov"}
	orgExplicit       = projectscope.Scope{KeyKind: projectscope.KeyKindOrganization}
)

func scopedResource(scope projectscope.Scope) (*ProjectScopedResource, *recordingClient) {
	raw := &recordingClient{missing: map[string]bool{}}
	r := &ProjectScopedResource{}
	r.client = sdkclient.Client{RawClient: raw, Scope: scope}
	return r, raw
}

// The schema of a resource with an identity (team_id) and a parent
// attribute, enough for every plan rule.
var testSchema = resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
	"id":         resourceschema.StringAttribute{Computed: true},
	"name":       resourceschema.StringAttribute{Required: true},
	"project_id": ProjectIDResourceAttribute(),
	"source_id":  resourceschema.StringAttribute{Optional: true},
	"team_id":    resourceschema.StringAttribute{Computed: true},
}}

// The schema of a resource whose v2 state records no project at all.
var childSchema = resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
	"project_id": ProjectIDResourceAttribute(),
	"source_id":  resourceschema.StringAttribute{Required: true},
}}

const unknown = "<unknown>"

// object builds a value of schema's type. An absent attribute is null, and
// the value unknown is an unknown.
func object(t *testing.T, schema resourceschema.Schema, attrs map[string]string) tftypes.Value {
	t.Helper()
	typ := schema.Type().TerraformType(t.Context())
	values := map[string]tftypes.Value{}
	for name := range schema.Attributes {
		switch v, ok := attrs[name]; {
		case !ok:
			values[name] = tftypes.NewValue(tftypes.String, nil)
		case v == unknown:
			values[name] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
		default:
			values[name] = tftypes.NewValue(tftypes.String, v)
		}
	}
	return tftypes.NewValue(typ, values)
}

type planCase struct {
	schema resourceschema.Schema
	// state is nil for a create.
	state  map[string]string
	config map[string]string
	// plan is what Terraform proposes before ModifyPlan.
	plan map[string]string
}

func modifyPlan(t *testing.T, r *ProjectScopedResource, c planCase) *resource.ModifyPlanResponse {
	t.Helper()
	stateRaw := tftypes.NewValue(c.schema.Type().TerraformType(t.Context()), nil)
	if c.state != nil {
		stateRaw = object(t, c.schema, c.state)
	}
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: c.schema, Raw: object(t, c.schema, c.config)},
		State:  tfsdk.State{Schema: c.schema, Raw: stateRaw},
		Plan:   tfsdk.Plan{Schema: c.schema, Raw: object(t, c.schema, c.plan)},
	}
	resp := &resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(t.Context(), req, resp)
	return resp
}

func plannedProject(t *testing.T, resp *resource.ModifyPlanResponse) types.String {
	t.Helper()
	var planned types.String
	if d := resp.Plan.GetAttribute(t.Context(), path.Root("project_id"), &planned); d.HasError() {
		t.Fatal(d)
	}
	return planned
}

func assertSummary(t *testing.T, diags diag.Diagnostics, want string) {
	t.Helper()
	if want == "" {
		if diags.HasError() {
			t.Fatalf("diagnostics = %v", diags)
		}
		return
	}
	if !diags.HasError() || diags[0].Summary() != want {
		t.Fatalf("diagnostics = %v, want %q", diags, want)
	}
}

func TestModifyPlan(t *testing.T) {
	existing := map[string]string{"id": "src_1", "name": "a", "project_id": "tm_old", "team_id": "tm_old"}
	v2State := map[string]string{"id": "src_1", "name": "a", "team_id": "tm_old"}

	cases := []struct {
		name        string
		scope       projectscope.Scope
		c           planCase
		wantErr     string
		wantProject string
		wantReplace bool
	}{
		{
			name:  "create, single-project: the provider's project is planned",
			scope: orgSingle,
			c: planCase{schema: testSchema, config: map[string]string{"name": "a"},
				plan: map[string]string{"name": "a", "id": unknown, "team_id": unknown, "project_id": unknown}},
			wantProject: "tm_prov",
		},
		{
			name:  "create, project key with unknown project: left to the API",
			scope: projectKeyUnknown,
			c: planCase{schema: testSchema, config: map[string]string{"name": "a"},
				plan: map[string]string{"name": "a", "id": unknown, "team_id": unknown, "project_id": unknown}},
			wantProject: unknown,
		},
		{
			name:  "create, explicit mode without project_id",
			scope: orgExplicit,
			c: planCase{schema: testSchema, config: map[string]string{"name": "a"},
				plan: map[string]string{"name": "a", "id": unknown, "team_id": unknown, "project_id": unknown}},
			wantErr: "Missing project_id",
		},
		{
			name:  "create, explicit mode, project_id known after apply",
			scope: orgExplicit,
			c: planCase{schema: testSchema, config: map[string]string{"name": "a", "project_id": unknown},
				plan: map[string]string{"name": "a", "id": unknown, "team_id": unknown, "project_id": unknown}},
			wantProject: unknown,
		},
		{
			name:  "create, single-project mode, another project_id",
			scope: projectKey,
			c: planCase{schema: testSchema, config: map[string]string{"name": "a", "project_id": "tm_other"},
				plan: map[string]string{"name": "a", "id": unknown, "team_id": unknown, "project_id": "tm_other"}},
			wantErr: "Project mismatch",
		},
		{
			name:        "no change: the plan is not touched",
			scope:       orgSingle,
			c:           planCase{schema: testSchema, state: map[string]string{"id": "src_1", "name": "a", "project_id": "tm_prov", "team_id": "tm_prov"}, config: map[string]string{"name": "a"}, plan: map[string]string{"id": "src_1", "name": "a", "project_id": "tm_prov", "team_id": "tm_prov"}},
			wantProject: "tm_prov",
		},
		{
			name:  "update: the stored project is planned instead of unknown",
			scope: projectKeyUnknown,
			c: planCase{schema: testSchema, state: existing, config: map[string]string{"name": "b"},
				plan: map[string]string{"id": "src_1", "name": "b", "project_id": unknown, "team_id": "tm_old"}},
			wantProject: "tm_old",
		},
		{
			name:  "v2 state, no change: no diff is added",
			scope: projectscope.Scope{KeyKind: projectscope.KeyKindProject, ProjectID: "tm_old"},
			c:     planCase{schema: testSchema, state: v2State, config: map[string]string{"name": "a"}, plan: v2State},
		},
		{
			name:  "v2 state, explicit mode with the same project: pinned by team_id",
			scope: orgExplicit,
			c: planCase{schema: testSchema, state: v2State, config: map[string]string{"name": "a", "project_id": "tm_old"},
				plan: map[string]string{"id": "src_1", "name": "a", "project_id": "tm_old", "team_id": "tm_old"}},
			wantProject: "tm_old",
		},
		{
			name:    "v2 state, provider now targets another project: pinned by team_id",
			scope:   projectKey,
			c:       planCase{schema: testSchema, state: v2State, config: map[string]string{"name": "a"}, plan: v2State},
			wantErr: "Project not reachable",
		},
		{
			name:        "org key, provider project changed: replace",
			scope:       orgSingle,
			c:           planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a"}, plan: existing},
			wantProject: "tm_prov", wantReplace: true,
		},
		{
			name:  "org key, resource project_id changed: replace",
			scope: orgExplicit,
			c: planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a", "project_id": "tm_new"},
				plan: map[string]string{"id": "src_1", "name": "a", "project_id": "tm_new", "team_id": "tm_old"}},
			wantProject: "tm_new", wantReplace: true,
		},
		{
			name:  "org key, project_id known after apply: replace",
			scope: orgExplicit,
			c: planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a", "project_id": unknown},
				plan: map[string]string{"id": "src_1", "name": "a", "project_id": unknown, "team_id": "tm_old"}},
			wantProject: unknown, wantReplace: true,
		},
		{
			name:    "org key, project_id removed in explicit mode",
			scope:   orgExplicit,
			c:       planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a"}, plan: existing},
			wantErr: "Missing project_id",
		},
		{
			name:    "project key swapped: error, never a replace",
			scope:   projectKey,
			c:       planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a"}, plan: existing},
			wantErr: "Project not reachable",
		},
		{
			name:  "child with v2 state, single-project: no diff is added",
			scope: projectKey,
			c: planCase{schema: childSchema, state: map[string]string{"source_id": "src_1"}, config: map[string]string{"source_id": "src_1"},
				plan: map[string]string{"source_id": "src_1"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := scopedResource(tc.scope)
			resp := modifyPlan(t, r, tc.c)

			assertSummary(t, resp.Diagnostics, tc.wantErr)
			if tc.wantErr != "" {
				if len(resp.RequiresReplace) != 0 {
					t.Error("an error must not also plan a replace")
				}
				return
			}
			planned := plannedProject(t, resp)
			switch {
			case tc.wantProject == unknown:
				if !planned.IsUnknown() {
					t.Errorf("planned project_id = %v, want unknown", planned)
				}
			case tc.wantProject == "":
				if !planned.IsNull() {
					t.Errorf("planned project_id = %v, want null", planned)
				}
			case planned.ValueString() != tc.wantProject:
				t.Errorf("planned project_id = %v, want %s", planned, tc.wantProject)
			}
			if replace := len(resp.RequiresReplace) > 0; replace != tc.wantReplace {
				t.Errorf("replace = %v, want %v", replace, tc.wantReplace)
			}
			if tc.wantReplace && !resp.RequiresReplace[0].Equal(path.Root("project_id")) {
				t.Errorf("replace path = %v", resp.RequiresReplace)
			}
		})
	}
}

func TestModifyPlan_destroyIsNotChecked(t *testing.T) {
	r, raw := scopedResource(projectKey)
	typ := testSchema.Type().TerraformType(t.Context())
	req := resource.ModifyPlanRequest{
		Config: tfsdk.Config{Schema: testSchema, Raw: tftypes.NewValue(typ, nil)},
		State:  tfsdk.State{Schema: testSchema, Raw: object(t, testSchema, map[string]string{"id": "src_1", "project_id": "tm_old"})},
		Plan:   tfsdk.Plan{Schema: testSchema, Raw: tftypes.NewValue(typ, nil)},
	}
	resp := &resource.ModifyPlanResponse{Plan: req.Plan}
	r.ModifyPlan(t.Context(), req, resp)
	assertSummary(t, resp.Diagnostics, "")
	if len(raw.paths) != 0 {
		t.Errorf("requests = %v", raw.paths)
	}
}

// A replace deletes before it creates, so the target project is checked
// first: a mistyped project must fail the plan, not the apply.
func TestModifyPlan_replaceChecksTheTargetProject(t *testing.T) {
	existing := map[string]string{"id": "src_1", "name": "a", "project_id": "tm_old", "team_id": "tm_old"}
	c := planCase{schema: testSchema, state: existing, config: map[string]string{"name": "a"}, plan: existing}

	r, raw := scopedResource(orgSingle)
	raw.missing["/"+sdkclient.APIVersion+"/projects/tm_prov"] = true
	resp := modifyPlan(t, r, c)

	assertSummary(t, resp.Diagnostics, "Project not accessible")
	if len(resp.RequiresReplace) != 0 {
		t.Error("an inaccessible target must not plan a replace")
	}
	if len(raw.projects) != 1 || raw.projects[0] != "" {
		t.Errorf("project check requests = %v, want one without a project header", raw.projects)
	}
}

func TestModifyPlan_parentAttribute(t *testing.T) {
	state := map[string]string{"source_id": "src_1", "project_id": "tm_old"}

	t.Run("same parent, another project: error, never a replace", func(t *testing.T) {
		r, _ := scopedResource(orgSingle)
		r.ParentAttribute = "source_id"
		resp := modifyPlan(t, r, planCase{schema: childSchema, state: state,
			config: map[string]string{"source_id": "src_1"}, plan: state})
		assertSummary(t, resp.Diagnostics, "Project mismatch")
		if len(resp.RequiresReplace) != 0 {
			t.Error("replace planned")
		}
	})

	for name, parent := range map[string]string{"new parent": "src_2", "parent known after apply": unknown} {
		t.Run(name+": moves in place", func(t *testing.T) {
			r, raw := scopedResource(orgSingle)
			r.ParentAttribute = "source_id"
			resp := modifyPlan(t, r, planCase{schema: childSchema, state: state,
				config: map[string]string{"source_id": parent},
				plan:   map[string]string{"source_id": parent, "project_id": unknown}})
			assertSummary(t, resp.Diagnostics, "")
			if got := plannedProject(t, resp); got.ValueString() != "tm_prov" {
				t.Errorf("planned project_id = %v, want tm_prov", got)
			}
			if len(resp.RequiresReplace) != 0 {
				t.Error("replace planned")
			}
			if len(raw.paths) != 0 {
				t.Errorf("requests = %v", raw.paths)
			}
		})
	}

	t.Run("project_id known after apply: no replace", func(t *testing.T) {
		r, _ := scopedResource(orgExplicit)
		r.ParentAttribute = "source_id"
		resp := modifyPlan(t, r, planCase{schema: childSchema, state: state,
			config: map[string]string{"source_id": "src_1", "project_id": unknown},
			plan:   map[string]string{"source_id": "src_1", "project_id": unknown}})
		assertSummary(t, resp.Diagnostics, "")
		if len(resp.RequiresReplace) != 0 {
			t.Error("replace planned")
		}
	})
}

func TestStoredProject(t *testing.T) {
	cases := []struct {
		projectID, teamID types.String
		want              string
	}{
		{types.StringValue("tm_p"), types.StringValue("tm_t"), "tm_p"},
		{types.StringNull(), types.StringValue("tm_t"), "tm_t"},
		{types.StringUnknown(), types.StringValue("tm_t"), "tm_t"},
		{types.StringNull(), types.StringNull(), ""},
	}
	for _, tc := range cases {
		if got := StoredProject(tc.projectID, tc.teamID); got != tc.want {
			t.Errorf("StoredProject(%v, %v) = %q, want %q", tc.projectID, tc.teamID, got, tc.want)
		}
	}
}

func sentProject(t *testing.T, client *sdkclient.Client, raw *recordingClient) string {
	t.Helper()
	if err := client.Do(t.Context(), "GET", "/sources/src_1", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	return raw.projects[len(raw.projects)-1]
}

func TestClientForState(t *testing.T) {
	cases := []struct {
		name    string
		scope   projectscope.Scope
		stored  string
		want    string
		wantErr string
	}{
		{"the stored project, not the provider's", orgSingle, "tm_old", "tm_old", ""},
		{"explicit mode", orgExplicit, "tm_old", "tm_old", ""},
		{"project key, own project", projectKey, "tm_key", "tm_key", ""},
		{"project key, project unknown: the stored project is sent", projectKeyUnknown, "tm_old", "tm_old", ""},
		{"project key swapped", projectKey, "tm_old", "", "Project not reachable"},
		{"nothing stored, single-project: the provider's project", orgSingle, "", "tm_prov", ""},
		{"nothing stored, project key with unknown project: no header", projectKeyUnknown, "", "", ""},
		{"nothing stored, explicit mode", orgExplicit, "", "", "Missing project_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, raw := scopedResource(tc.scope)
			var diags diag.Diagnostics
			client, ok := r.ClientForState(tc.stored, &diags)
			assertSummary(t, diags, tc.wantErr)
			if ok != (tc.wantErr == "") {
				t.Fatalf("ok = %v", ok)
			}
			if !ok {
				return
			}
			if got := sentProject(t, client, raw); got != tc.want {
				t.Errorf("project header = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClientForCreate(t *testing.T) {
	cases := []struct {
		name    string
		scope   projectscope.Scope
		planned types.String
		want    string
		wantErr string
	}{
		{"single-project", orgSingle, types.StringValue("tm_prov"), "tm_prov", ""},
		{"single-project, project_id resolved at apply to another project", orgSingle, types.StringValue("tm_other"), "", "Project mismatch"},
		{"explicit mode", orgExplicit, types.StringValue("tm_res"), "tm_res", ""},
		{"explicit mode, nothing planned", orgExplicit, types.StringUnknown(), "", "Missing project_id"},
		{"project key with unknown project: no header", projectKeyUnknown, types.StringUnknown(), "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, raw := scopedResource(tc.scope)
			var diags diag.Diagnostics
			client, ok := r.ClientForCreate(tc.planned, &diags)
			assertSummary(t, diags, tc.wantErr)
			if !ok {
				return
			}
			if got := sentProject(t, client, raw); got != tc.want {
				t.Errorf("project header = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestImportProjectState(t *testing.T) {
	cases := []struct {
		name        string
		scope       projectscope.Scope
		id          string
		wantErr     string
		wantID      string
		wantProject string
	}{
		{"bare id, single-project", projectKey, "src_1", "", "src_1", ""},
		{"project and id", orgExplicit, "tm_res/src_1", "", "src_1", "tm_res"},
		{"the provider's project and id", projectKey, "tm_key/src_1", "", "src_1", "tm_key"},
		{"bare id, explicit mode", orgExplicit, "src_1", "Missing project in import ID", "", ""},
		{"another project, single-project", projectKey, "tm_other/src_1", "Project mismatch", "", ""},
		{"malformed", projectKey, "a/b/c", "Invalid import ID", "", ""},
		{"empty project", projectKey, "/src_1", "Invalid import ID", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := scopedResource(tc.scope)
			resp := &resource.ImportStateResponse{State: tfsdk.State{Schema: testSchema, Raw: object(t, testSchema, nil)}}
			r.ImportProjectState(t.Context(), "id", resource.ImportStateRequest{ID: tc.id}, resp)

			assertSummary(t, resp.Diagnostics, tc.wantErr)
			var id, project types.String
			resp.State.GetAttribute(t.Context(), path.Root("id"), &id)
			resp.State.GetAttribute(t.Context(), path.Root("project_id"), &project)
			if id.ValueString() != tc.wantID || project.ValueString() != tc.wantProject {
				t.Errorf("state = (%v, %v), want (%q, %q)", id, project, tc.wantID, tc.wantProject)
			}
		})
	}
}

func TestProjectDiagnostics(t *testing.T) {
	cases := []struct {
		err     error
		summary string
		detail  []string
	}{
		{projectscope.ErrProjectRequired, "Missing project_id", []string{"organization API key", "project_id is required"}},
		{&projectscope.MismatchError{Provider: "tm_prov", Configured: "tm_res"}, "Project mismatch", []string{"tm_prov", "tm_res"}},
		{&projectscope.UnreachableError{Stored: "tm_old", Target: "tm_new"}, "Project not reachable", []string{"tm_old", "tm_new", "terraform state rm", "removed block", "organization API key"}},
		{&sdkclient.ProjectAccessError{ProjectID: "tm_x", Status: 404}, "Project not accessible", []string{"tm_x"}},
		{io.ErrUnexpectedEOF, "Error resolving project", []string{io.ErrUnexpectedEOF.Error()}},
	}
	for _, tc := range cases {
		diags := ProjectDiagnostics(tc.err)
		if len(diags) != 1 || diags[0].Summary() != tc.summary {
			t.Errorf("%v: diagnostics = %v, want %q", tc.err, diags, tc.summary)
			continue
		}
		for _, want := range tc.detail {
			if !strings.Contains(diags[0].Detail(), want) {
				t.Errorf("%v: detail %q does not contain %q", tc.err, diags[0].Detail(), want)
			}
		}
	}
}
