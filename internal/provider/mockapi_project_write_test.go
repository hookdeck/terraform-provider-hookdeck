package provider_test

import (
	"context"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// protoProject drives a project resource through the provider protocol, so a
// write can run without the refresh a Terraform plan or destroy starts with.
// plugin-testing cannot skip it, and with the refresh a key that cannot reach
// the project fails there, before the write.
type protoProject struct {
	t        *testing.T
	server   tfprotov6.ProviderServer
	typeName string
	typ      tftypes.Object
}

func newProtoProject(t *testing.T, typeName string) *protoProject {
	t.Helper()
	ctx := context.Background()
	server, err := acc.ProviderFactories["hookdeck"]()
	if err != nil {
		t.Fatal(err)
	}
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	providerType := objectType(t, schemas.Provider.ValueType())
	config, err := tfprotov6.NewDynamicValue(providerType, nullObject(providerType, nil))
	if err != nil {
		t.Fatal(err)
	}
	// The key comes from the environment (acc.UseAPIKey).
	configured, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: &config})
	if err != nil {
		t.Fatal(err)
	}
	failOnError(t, configured.Diagnostics)
	schema, ok := schemas.ResourceSchemas[typeName]
	if !ok {
		t.Fatalf("no resource %s", typeName)
	}
	return &protoProject{t: t, server: server, typeName: typeName, typ: objectType(t, schema.ValueType())}
}

func objectType(t *testing.T, typ tftypes.Type) tftypes.Object {
	t.Helper()
	object, ok := typ.(tftypes.Object)
	if !ok {
		t.Fatalf("schema type is %T, want an object", typ)
	}
	return object
}

func nullObject(typ tftypes.Object, set map[string]tftypes.Value) tftypes.Value {
	values := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	for name, v := range set {
		values[name] = v
	}
	return tftypes.NewValue(typ, values)
}

func failOnError(t *testing.T, diags []*tfprotov6.Diagnostic) {
	t.Helper()
	for _, d := range diags {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("%s: %s", d.Summary, d.Detail)
		}
	}
}

// state is the resource as Terraform holds it after an apply.
func (p *protoProject) state(id, name string) tftypes.Value {
	return nullObject(p.typ, map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, id),
		"name":            tftypes.NewValue(tftypes.String, name),
		"type":            tftypes.NewValue(tftypes.String, p.apiType()),
		"organization_id": tftypes.NewValue(tftypes.String, mockOrg),
		"created_at":      tftypes.NewValue(tftypes.String, mockTime),
		"updated_at":      tftypes.NewValue(tftypes.String, mockTime),
	})
}

func (p *protoProject) apiType() string { return apiTypeOf(p.typeName) }

func apiTypeOf(typeName string) string {
	if typeName == "hookdeck_outpost_project" {
		return "outpost"
	}
	return "event_gateway"
}

// apply sends one ApplyResourceChange and returns the error diagnostics and
// the state Terraform keeps afterwards.
func (p *protoProject) apply(prior, planned tftypes.Value) ([]*tfprotov6.Diagnostic, tftypes.Value) {
	p.t.Helper()
	dynamic := func(v tftypes.Value) *tfprotov6.DynamicValue {
		d, err := tfprotov6.NewDynamicValue(p.typ, v)
		if err != nil {
			p.t.Fatal(err)
		}
		return &d
	}
	resp, err := p.server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{
		TypeName:     p.typeName,
		PriorState:   dynamic(prior),
		PlannedState: dynamic(planned),
		Config:       dynamic(planned),
	})
	if err != nil {
		p.t.Fatal(err)
	}
	var errs []*tfprotov6.Diagnostic
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			errs = append(errs, d)
		}
	}
	after, err := resp.NewState.Unmarshal(p.typ)
	if err != nil {
		p.t.Fatal(err)
	}
	return errs, after
}

func (p *protoProject) delete(id, name string) ([]*tfprotov6.Diagnostic, tftypes.Value) {
	p.t.Helper()
	return p.apply(p.state(id, name), tftypes.NewValue(p.typ, nil))
}

func (p *protoProject) rename(id, name, newName string) ([]*tfprotov6.Diagnostic, tftypes.Value) {
	p.t.Helper()
	return p.apply(p.state(id, name), p.state(id, newName))
}

var projectResources = []string{"hookdeck_gateway_project", "hookdeck_outpost_project"}

func expectWriteError(t *testing.T, diags []*tfprotov6.Diagnostic, summary string) {
	t.Helper()
	if len(diags) == 0 {
		t.Fatalf("no error, want %q", summary)
	}
	if diags[0].Summary != summary {
		t.Errorf("error %q, want %q", diags[0].Summary, summary)
	}
	if !regexp.MustCompile(`(?s)cannot\s+access\s+project.*Use\s+an\s+API\s+key`).MatchString(diags[0].Detail) {
		t.Errorf("detail = %q", diags[0].Detail)
	}
}

// expectDeleteRefused checks a destroy the API refused. An Outpost project
// removes its portal custom domain first, so the refusal can come from either
// call.
func expectDeleteRefused(t *testing.T, typeName string, diags []*tfprotov6.Diagnostic) {
	t.Helper()
	if typeName != "hookdeck_outpost_project" {
		expectWriteError(t, diags, "Error deleting project")
		return
	}
	if len(diags) == 0 {
		t.Fatal("no error, want a refused delete")
	}
	if s := diags[0].Summary; s != "Error deleting project" && s != "Error deleting portal custom domain" {
		t.Errorf("error %q, want a refused delete", s)
	}
	if !regexp.MustCompile(`(?s)cannot\s+access\s+project.*(Use\s+an|Check\s+that\s+the)\s+API\s+key`).MatchString(diags[0].Detail) {
		t.Errorf("detail = %q", diags[0].Detail)
	}
}

// A destroy that skips the refresh, with a key that cannot reach the project,
// fails and keeps the project and its state.
func TestAccMock_ProjectDelete_UnreachableKeepsProject(t *testing.T) {
	acc.SkipUnlessAcc(t)
	keys := map[string]mockKey{
		"another organization":      {org: mockOtherOrg},
		"key limited to others":     {org: mockOrg, grants: map[string]bool{"tm_a": true}},
		"key without projects.read": {org: mockOrg, noProjectsRead: true},
	}
	for _, typeName := range projectResources {
		for name, key := range keys {
			t.Run(typeName+"/"+name, func(t *testing.T) {
				m := mockOrganization(t)
				m.addProject("tm_target", "target", apiTypeOf(typeName))
				m.addKey("hd_org_mock_unreachable", key)
				acc.UseAPIKey(t, "hd_org_mock_unreachable")
				p := newProtoProject(t, typeName)

				diags, after := p.delete("tm_target", "target")

				expectDeleteRefused(t, typeName, diags)
				if !after.Equal(p.state("tm_target", "target")) {
					t.Errorf("state = %v, want it unchanged", after)
				}
				if got := m.projects["tm_target"]; got == nil {
					t.Error("project was deleted")
				}
			})
		}
	}
}

// A project the API says is gone is deleted already: a destroy that skips the
// refresh succeeds, whichever 404 the API answers.
func TestAccMock_ProjectDelete_AlreadyDeletedSucceeds(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, typeName := range projectResources {
		t.Run(typeName+"/unknown project", func(t *testing.T) {
			m := mockOrganization(t)
			acc.UseAPIKey(t, mockOrgKey)
			p := newProtoProject(t, typeName)

			diags, after := p.delete("tm_gone", "gone")

			if len(diags) != 0 || !after.IsNull() {
				t.Errorf("diagnostics = %v, state null = %v", diags, after.IsNull())
			}
			if got := countRequests(m, "DELETE", "/projects/tm_gone"); got != 1 {
				t.Errorf("%d DELETE requests, want 1", got)
			}
		})
		t.Run(typeName+"/just deleted", func(t *testing.T) {
			m := mockOrganization(t)
			m.addProject("tm_target", "target", apiTypeOf(typeName))
			m.faults["DELETE projects"] = 404
			acc.UseAPIKey(t, mockOrgKey)
			p := newProtoProject(t, typeName)

			diags, after := p.delete("tm_target", "target")

			if len(diags) != 0 || !after.IsNull() {
				t.Errorf("diagnostics = %v, state null = %v", diags, after.IsNull())
			}
		})
	}
}

// Renaming a project the key cannot reach fails, and the project keeps its
// name.
func TestAccMock_ProjectUpdate_UnreachableFails(t *testing.T) {
	acc.SkipUnlessAcc(t)
	keys := map[string]mockKey{
		"another organization":  {org: mockOtherOrg},
		"key limited to others": {org: mockOrg, grants: map[string]bool{"tm_a": true}},
	}
	for _, typeName := range projectResources {
		for name, key := range keys {
			t.Run(typeName+"/"+name, func(t *testing.T) {
				m := mockOrganization(t)
				m.addProject("tm_target", "target", apiTypeOf(typeName))
				m.addKey("hd_org_mock_unreachable", key)
				acc.UseAPIKey(t, "hd_org_mock_unreachable")
				p := newProtoProject(t, typeName)

				diags, after := p.rename("tm_target", "target", "renamed")

				expectWriteError(t, diags, "Error updating project")
				if got := m.projects["tm_target"].Name; got != "target" {
					t.Errorf("project name = %q", got)
				}
				if !after.Equal(p.state("tm_target", "target")) {
					t.Errorf("state = %v, want it unchanged", after)
				}
			})
		}
		t.Run(typeName+"/missing project", func(t *testing.T) {
			mockOrganization(t)
			acc.UseAPIKey(t, mockOrgKey)
			p := newProtoProject(t, typeName)

			diags, _ := p.rename("tm_gone", "gone", "renamed")

			if len(diags) == 0 || diags[0].Summary != "Error updating project" || diags[0].Detail != "Project tm_gone no longer exists." {
				t.Errorf("diagnostics = %v", diags)
			}
		})
	}
}
