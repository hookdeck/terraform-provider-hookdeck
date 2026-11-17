package connection

import (
	"context"
	"testing"
)

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

func TestUpdatePayloadRules(t *testing.T) {
	one := []rule{{DelayRule: &delayRule{}}}
	tests := []struct {
		name       string
		rules      []rule
		prior      []rule
		wantSent   bool
		wantLength int
	}{
		{"never had rules", nil, nil, false, 0},
		{"prior empty list", nil, []rule{}, false, 0},
		{"removed after set", nil, one, true, 0},
		{"empty list", []rule{}, one, true, 0},
		{"empty list, none before", []rule{}, nil, true, 0},
		{"set", one, nil, true, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := connectionResourceModel{Rules: tt.rules}
			payload, err := m.toUpdatePayload(context.Background(), tt.prior)
			if err != nil {
				t.Fatal(err)
			}
			sent, present := payload["rules"]
			if present != tt.wantSent {
				t.Fatalf("rules sent = %v, want %v (%v)", present, tt.wantSent, sent)
			}
			if present {
				if list, ok := sent.([]interface{}); !ok || len(list) != tt.wantLength {
					t.Fatalf("rules = %#v, want a list of %d", sent, tt.wantLength)
				}
			}
		})
	}
}
