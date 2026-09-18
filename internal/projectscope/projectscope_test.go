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

func TestResolve(t *testing.T) {
	keyProject := func(id string) func() (string, error) {
		return func() (string, error) { return id, nil }
	}
	neverCalled := func() (string, error) {
		panic("keyProject must not be called")
	}

	cases := []struct {
		name       string
		kind       KeyKind
		resource   string
		provider   string
		keyProject func() (string, error)
		want       string
		wantErr    error
	}{
		{"org: resource wins", KeyKindOrganization, "tm_res", "tm_prov", neverCalled, "tm_res", nil},
		{"org: provider default", KeyKindOrganization, "", "tm_prov", neverCalled, "tm_prov", nil},
		{"org: nothing", KeyKindOrganization, "", "", neverCalled, "", ErrProjectRequired},
		{"project: nothing set, no header", KeyKindProject, "", "", neverCalled, "", nil},
		{"project: resource matches", KeyKindProject, "tm_key", "", keyProject("tm_key"), "tm_key", nil},
		{"project: resource mismatch", KeyKindProject, "tm_other", "", keyProject("tm_key"), "", ErrProjectMismatch},
		{"project: provider matches", KeyKindProject, "", "tm_key", keyProject("tm_key"), "tm_key", nil},
		{"project: provider mismatch", KeyKindProject, "", "tm_other", keyProject("tm_key"), "", ErrProjectMismatch},
		{"project: resource beats provider", KeyKindProject, "tm_key", "tm_other", keyProject("tm_key"), "tm_key", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(tc.kind, tc.resource, tc.provider, tc.keyProject)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolve_keyProjectError(t *testing.T) {
	boom := errors.New("boom")
	_, err := Resolve(KeyKindProject, "tm_x", "", func() (string, error) { return "", boom })
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
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
