package source

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
)

func TestRefreshSetsConfig(t *testing.T) {
	tests := []struct {
		name   string
		config interface{}
		want   jsontypes.Normalized
	}{
		{
			name:   "object",
			config: map[string]interface{}{"auth": map[string]interface{}{"type": "header"}},
			want:   jsontypes.NewNormalizedValue(`{"auth":{"type":"header"}}`),
		},
		{
			name:   "empty object",
			config: map[string]interface{}{},
			want:   jsontypes.NewNormalizedValue(`{}`),
		},
		{
			name:   "null",
			config: nil,
			want:   jsontypes.NewNormalizedNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := sourceResourceModel{}
			diags := m.Refresh(map[string]interface{}{
				"created_at": "2026-01-01T00:00:00Z",
				"id":         "source-test-id",
				"name":       "test",
				"team_id":    "tm_test",
				"type":       "HTTP",
				"updated_at": "2026-01-01T00:00:00Z",
				"url":        "https://example.invalid/source-test-id",
				"config":     tt.config,
			})
			if diags.HasError() {
				t.Fatal(diags)
			}
			if !reflect.DeepEqual(m.Config, tt.want) {
				t.Fatalf("config = %#v, want %#v", m.Config, tt.want)
			}
		})
	}
}
