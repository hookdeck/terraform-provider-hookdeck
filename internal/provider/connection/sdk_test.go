package connection

import "testing"

func refreshResponse(rules []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"created_at": "2026-01-01T00:00:00Z",
		"id":         "web_test",
		"team_id":    "tm_test",
		"updated_at": "2026-01-01T00:00:00Z",
		"rules":      rules,
	}
}

func TestRefreshClearsRulesRemovedOutsideTerraform(t *testing.T) {
	m := connectionResourceModel{Rules: []rule{{DelayRule: &delayRule{}}}}
	if diags := m.Refresh(refreshResponse([]interface{}{})); diags.HasError() {
		t.Fatal(diags)
	}
	if m.Rules == nil || len(m.Rules) != 0 {
		t.Fatalf("expected an empty, non-nil rules list, got %#v", m.Rules)
	}
}

func TestRefreshKeepsConfiguredEmptyRulesNonNil(t *testing.T) {
	m := connectionResourceModel{Rules: []rule{}}
	if diags := m.Refresh(refreshResponse([]interface{}{})); diags.HasError() {
		t.Fatal(diags)
	}
	if m.Rules == nil {
		t.Fatal("expected rules to stay an empty list, got nil")
	}
}

func TestRefreshKeepsUnconfiguredRulesNil(t *testing.T) {
	m := connectionResourceModel{}
	if diags := m.Refresh(refreshResponse([]interface{}{})); diags.HasError() {
		t.Fatal(diags)
	}
	if m.Rules != nil {
		t.Fatalf("expected rules to stay nil, got %#v", m.Rules)
	}
}
