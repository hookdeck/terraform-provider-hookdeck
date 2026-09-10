package connection

import "testing"

func TestRefreshClearsRulesWhenAPIReturnsEmptyList(t *testing.T) {
	m := connectionResourceModel{Rules: []rule{{DelayRule: &delayRule{}}}}
	diags := m.Refresh(map[string]interface{}{
		"created_at": "2026-01-01T00:00:00Z",
		"id":         "connection-test-id",
		"team_id":    "tm_test",
		"updated_at": "2026-01-01T00:00:00Z",
		"rules":      []interface{}{},
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	if len(m.Rules) != 0 {
		t.Fatalf("rules = %#v, want empty list", m.Rules)
	}
}
