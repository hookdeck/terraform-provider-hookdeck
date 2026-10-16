package projectscope

import (
	"errors"
	"testing"
)

func TestKindOfKey(t *testing.T) {
	cases := map[string]KeyKind{
		"hd_org_a1b2c3d4e5f6":  KeyKindOrganization,
		"hd_a1b2c3d4e5f6":      KeyKindProject,
		"a1b2c3d4e5f6":         KeyKindProject,
		"":                     KeyKindProject,
		"HD_ORG_a1b2c3d4e5f6":  KeyKindProject, // prefix is case-sensitive
		"hd_organization_key1": KeyKindProject,
	}
	for key, want := range cases {
		if got := KindOfKey(key); got != want {
			t.Errorf("KindOfKey(%q) = %v, want %v", key, got, want)
		}
	}
}

var (
	projectKey        = Scope{KeyKind: KeyKindProject, ProjectID: "tm_key"}
	projectKeyUnknown = Scope{KeyKind: KeyKindProject}
	orgSingle         = Scope{KeyKind: KeyKindOrganization, ProjectID: "tm_prov"}
	orgExplicit       = Scope{KeyKind: KeyKindOrganization}
)

func TestScope_SingleProject(t *testing.T) {
	cases := map[Scope]bool{
		projectKey:        true,
		projectKeyUnknown: true,
		orgSingle:         true,
		orgExplicit:       false,
	}
	for scope, want := range cases {
		if got := scope.SingleProject(); got != want {
			t.Errorf("%+v SingleProject() = %v, want %v", scope, got, want)
		}
	}
}

func TestScope_Target(t *testing.T) {
	cases := []struct {
		name       string
		scope      Scope
		configured string
		want       string
		wantErr    string
	}{
		{"project key: unset uses the key's project", projectKey, "", "tm_key", ""},
		{"project key: equal is fine", projectKey, "tm_key", "tm_key", ""},
		{"project key: different is a mismatch", projectKey, "tm_other", "", "mismatch"},
		{"project key, project unknown: unset stays unknown", projectKeyUnknown, "", "", ""},
		{"project key, project unknown: configured is used", projectKeyUnknown, "tm_res", "tm_res", ""},
		{"org single: unset uses the provider's project", orgSingle, "", "tm_prov", ""},
		{"org single: equal is fine", orgSingle, "tm_prov", "tm_prov", ""},
		{"org single: different is a mismatch, not an override", orgSingle, "tm_res", "", "mismatch"},
		{"org explicit: configured is used", orgExplicit, "tm_res", "tm_res", ""},
		{"org explicit: unset is an error", orgExplicit, "", "", "required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.scope.Target(tc.configured)
			assertErr(t, err, tc.wantErr)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestScope_Target_mismatchNamesBothProjects(t *testing.T) {
	_, err := orgSingle.Target("tm_res")
	var mismatch *MismatchError
	if !errors.As(err, &mismatch) || mismatch.Provider != "tm_prov" || mismatch.Configured != "tm_res" {
		t.Fatalf("err = %#v", err)
	}
}

func TestScope_Plan(t *testing.T) {
	cases := []struct {
		name       string
		scope      Scope
		configured string
		stored     string
		want       string
		wantAction Action
		wantErr    string
	}{
		{"new resource takes the target", orgSingle, "", "", "tm_prov", ActionKeep, ""},
		{"new resource, explicit", orgExplicit, "tm_res", "", "tm_res", ActionKeep, ""},
		{"new resource, explicit, unset", orgExplicit, "", "", "", ActionKeep, "required"},
		{"same project, unset", projectKey, "", "tm_key", "tm_key", ActionKeep, ""},
		{"same project, set", projectKey, "tm_key", "tm_key", "tm_key", ActionKeep, ""},
		{"project key, project unknown: stored is kept", projectKeyUnknown, "", "tm_old", "tm_old", ActionKeep, ""},
		{"project key swapped: unreachable, never a replace", projectKey, "", "tm_old", "", ActionKeep, "unreachable"},
		{"project key, project unknown, other project_id: unreachable", projectKeyUnknown, "tm_new", "tm_old", "", ActionKeep, "unreachable"},
		{"project key, different project_id: mismatch", projectKey, "tm_other", "tm_key", "", ActionKeep, "mismatch"},
		{"org single: provider project changed", orgSingle, "", "tm_old", "tm_prov", ActionReplace, ""},
		{"org single: unchanged", orgSingle, "", "tm_prov", "tm_prov", ActionKeep, ""},
		{"org explicit: project_id changed", orgExplicit, "tm_new", "tm_old", "tm_new", ActionReplace, ""},
		{"org explicit: unchanged", orgExplicit, "tm_old", "tm_old", "tm_old", ActionKeep, ""},
		{"org explicit: project_id removed", orgExplicit, "", "tm_old", "", ActionKeep, "required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, action, err := tc.scope.Plan(tc.configured, tc.stored)
			assertErr(t, err, tc.wantErr)
			if got != tc.want || action != tc.wantAction {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, action, tc.want, tc.wantAction)
			}
		})
	}
}

func TestScope_Plan_unreachableNamesBothProjects(t *testing.T) {
	_, _, err := projectKey.Plan("", "tm_old")
	var unreachable *UnreachableError
	if !errors.As(err, &unreachable) || unreachable.Stored != "tm_old" || unreachable.Target != "tm_key" {
		t.Fatalf("err = %#v", err)
	}
}

func TestScope_Reach(t *testing.T) {
	cases := []struct {
		name    string
		scope   Scope
		stored  string
		wantErr string
	}{
		{"project key, own project", projectKey, "tm_key", ""},
		{"project key, other project", projectKey, "tm_old", "unreachable"},
		{"project key, project unknown: left to the API", projectKeyUnknown, "tm_old", ""},
		{"org key reaches any project", orgSingle, "tm_old", ""},
		{"org key, explicit", orgExplicit, "tm_old", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertErr(t, tc.scope.Reach(tc.stored), tc.wantErr)
		})
	}
}

func assertErr(t *testing.T, err error, want string) {
	t.Helper()
	var mismatch *MismatchError
	var unreachable *UnreachableError
	got := ""
	switch {
	case err == nil:
	case errors.Is(err, ErrProjectRequired):
		got = "required"
	case errors.As(err, &mismatch):
		got = "mismatch"
	case errors.As(err, &unreachable):
		got = "unreachable"
	default:
		got = err.Error()
	}
	if got != want {
		t.Fatalf("err = %q (%v), want %q", got, err, want)
	}
}

func TestParseImportID(t *testing.T) {
	cases := []struct {
		in, project, id string
		wantErr         bool
	}{
		{"src_abc", "", "src_abc", false},
		{"tm_123/src_abc", "tm_123", "src_abc", false},
		{"tm_123/", "", "", true},
		{"/src_abc", "", "", true},
		{"a/b/c", "", "", true},
		{"", "", "", true},
	}
	for _, tc := range cases {
		project, id, err := ParseImportID(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseImportID(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
			continue
		}
		if project != tc.project || id != tc.id {
			t.Errorf("ParseImportID(%q) = (%q, %q), want (%q, %q)", tc.in, project, id, tc.project, tc.id)
		}
	}
}
