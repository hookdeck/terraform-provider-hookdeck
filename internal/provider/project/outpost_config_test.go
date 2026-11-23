package project

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func strs(kv ...string) map[string]types.String {
	out := map[string]types.String{}
	for i := 0; i < len(kv); i += 2 {
		out[kv[i]] = types.StringValue(kv[i+1])
	}
	return out
}

func messages(diags diag.Diagnostics, severity diag.Severity) string {
	var out []string
	for _, d := range diags {
		if d.Severity() == severity {
			out = append(out, d.Summary()+": "+d.Detail())
		}
	}
	return strings.Join(out, "\n")
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]types.String
		sensitive map[string]types.String
		wantError []string
	}{
		{name: "any key in either map", config: strs("TOPICS", "a,b", "NEW_FUTURE_KEY", "x"), sensitive: strs("OTEL_EXPORTER_OTLP_HEADERS", "k=v", "TOPICS_ALLOW_WILDCARDS", "true")},
		{name: "credentials in config are the user's call", config: strs("OTEL_EXPORTER_OTLP_ENDPOINT", "https://user:token@otel.example.com", "DESTINATIONS_PROXY_URL", "http://u:p@proxy")},
		{name: "key in both maps", config: strs("TOPICS", "a"), sensitive: strs("TOPICS", "a"), wantError: []string{"TOPICS", "both"}},
		{name: "null value in config", config: map[string]types.String{"TOPICS": types.StringNull()}, wantError: []string{"TOPICS", "Remove the key"}},
		{name: "null value in sensitive_config", sensitive: map[string]types.String{"OTEL_EXPORTER_OTLP_HEADERS": types.StringNull()}, wantError: []string{"OTEL_EXPORTER_OTLP_HEADERS", "Remove the key"}},
		{name: "null and unknown maps", config: nil, sensitive: nil},
		{name: "unknown value", config: map[string]types.String{"TOPICS": types.StringUnknown()}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateConfig(tc.config, tc.sensitive)
			errs, warns := messages(diags, diag.SeverityError), messages(diags, diag.SeverityWarning)
			if warns != "" {
				t.Errorf("unexpected warnings: %q", warns)
			}
			if (errs == "") != (len(tc.wantError) == 0) {
				t.Fatalf("errors = %q, want %v", errs, tc.wantError)
			}
			for _, want := range tc.wantError {
				if !strings.Contains(errs, want) {
					t.Errorf("error %q does not contain %q", errs, want)
				}
			}
		})
	}
}

func ptr(s string) *string { return &s }

func TestPlanPatch(t *testing.T) {
	live := map[string]*string{
		"TOPICS":                 ptr("a,b"),
		"RETRY_INTERVAL_SECONDS": ptr("30"),
		"MAX_RETRY_LIMIT":        ptr("10"),
		"DESTINATIONS_PROXY_URL": nil,
		"PORTAL_LOGO":            nil,
	}
	tests := []struct {
		name           string
		desired, prior map[string]string
		want           map[string]*string
	}{
		{name: "changed key only", desired: map[string]string{"TOPICS": "a,b", "RETRY_INTERVAL_SECONDS": "60"}, prior: map[string]string{"TOPICS": "a,b", "RETRY_INTERVAL_SECONDS": "30"},
			want: map[string]*string{"RETRY_INTERVAL_SECONDS": ptr("60")}},
		{name: "removed key is null", desired: map[string]string{"TOPICS": "a,b"}, prior: map[string]string{"TOPICS": "a,b", "RETRY_INTERVAL_SECONDS": "30"},
			want: map[string]*string{"RETRY_INTERVAL_SECONDS": nil}},
		{name: "removed key already unset live", desired: map[string]string{}, prior: map[string]string{"PORTAL_LOGO": "x"}, want: map[string]*string{}},
		{name: "key moved between maps", desired: map[string]string{"TOPICS": "a,b"}, prior: map[string]string{"TOPICS": "a,b"}, want: map[string]*string{}},
		{name: "equal to live after import", desired: map[string]string{"TOPICS": "a,b", "MAX_RETRY_LIMIT": "10"}, prior: nil, want: map[string]*string{}},
		{name: "empty string against a default is sent", desired: map[string]string{"TOPICS": ""}, prior: nil, want: map[string]*string{"TOPICS": ptr("")}},
		{name: "empty string against null is sent for other keys", desired: map[string]string{"PORTAL_LOGO": ""}, prior: nil, want: map[string]*string{"PORTAL_LOGO": ptr("")}},
		{name: "new key", desired: map[string]string{"OTEL_SERVICE_NAME": "svc"}, prior: nil, want: map[string]*string{"OTEL_SERVICE_NAME": ptr("svc")}},
		{name: "empty", desired: nil, prior: nil, want: map[string]*string{}},
		{name: "value moved from config to sensitive_config keeps its value", desired: map[string]string{"TOPICS": "a,b", "MAX_RETRY_LIMIT": "10"}, prior: map[string]string{"TOPICS": "a,b", "MAX_RETRY_LIMIT": "10"}, want: map[string]*string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := planPatch(tc.desired, tc.prior, live)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("patch = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRefreshManaged(t *testing.T) {
	live := map[string]*string{
		"TOPICS":                 ptr("a,b,c"),
		"RETRY_INTERVAL_SECONDS": nil,
		"DESTINATIONS_PROXY_URL": nil,
		"UNMANAGED":              ptr("x"),
	}
	managed := map[string]string{
		"TOPICS":                 "a",
		"RETRY_INTERVAL_SECONDS": "30",
		"DESTINATIONS_PROXY_URL": "",
		"MISSING_LIVE":           "1",
	}
	want := map[string]string{"TOPICS": "a,b,c", "DESTINATIONS_PROXY_URL": ""}
	if got := refreshManaged(managed, live); !reflect.DeepEqual(got, want) {
		t.Errorf("managed = %v, want %v", got, want)
	}
}

// Both maps are merged into one object for the API, and a key that moves from
// one map to the other keeps its value: nothing is sent, and certainly no null.
func TestManagedMergesBothMaps(t *testing.T) {
	ctx := context.Background()
	mapOf := func(kv ...string) types.Map {
		m := map[string]string{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		v, _ := types.MapValueFrom(ctx, types.StringType, m)
		return v
	}
	before := outpostProjectModel{Config: mapOf("TOPICS", "a", "PORTAL_LOGO", "x"), SensitiveConfig: mapOf("OTEL_EXPORTER_OTLP_HEADERS", "k=v")}
	after := outpostProjectModel{Config: mapOf("PORTAL_LOGO", "x"), SensitiveConfig: mapOf("OTEL_EXPORTER_OTLP_HEADERS", "k=v", "TOPICS", "a")}

	desired, diags := after.managed(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	prior, _ := before.managed(ctx)
	want := map[string]string{"TOPICS": "a", "PORTAL_LOGO": "x", "OTEL_EXPORTER_OTLP_HEADERS": "k=v"}
	if !reflect.DeepEqual(desired, want) || !reflect.DeepEqual(prior, want) {
		t.Errorf("desired = %v, prior = %v, want both %v", desired, prior, want)
	}
	live := map[string]*string{"TOPICS": ptr("a"), "PORTAL_LOGO": ptr("x"), "OTEL_EXPORTER_OTLP_HEADERS": ptr("k=v")}
	if patch := planPatch(desired, prior, live); len(patch) != 0 {
		t.Errorf("patch = %v, want none", patch)
	}
}

// Refresh puts each live value in the map that holds the key; keys the user
// does not hold are not added to either map.
func TestSetManagedKeepsKeysByMembership(t *testing.T) {
	ctx := context.Background()
	config, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"TOPICS": "a"})
	sensitive, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"OTEL_EXPORTER_OTLP_HEADERS": "k=v"})
	m := outpostProjectModel{Config: config, SensitiveConfig: sensitive}
	diags := m.setManaged(ctx, map[string]string{"TOPICS": "b", "OTEL_EXPORTER_OTLP_HEADERS": "k=w", "UNMANAGED": "x"})
	if diags.HasError() {
		t.Fatal(diags)
	}
	gotConfig, _ := stringMap(ctx, m.Config)
	gotSensitive, _ := stringMap(ctx, m.SensitiveConfig)
	if !reflect.DeepEqual(gotConfig, map[string]string{"TOPICS": "b"}) || !reflect.DeepEqual(gotSensitive, map[string]string{"OTEL_EXPORTER_OTLP_HEADERS": "k=w"}) {
		t.Errorf("config = %v, sensitive_config = %v", gotConfig, gotSensitive)
	}
}
