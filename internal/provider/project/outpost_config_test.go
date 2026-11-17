package project

import (
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
		wantWarn  []string
	}{
		{name: "known keys pass", config: strs("TOPICS", "a,b", "RETRY_INTERVAL_SECONDS", "5"), sensitive: strs("OTEL_EXPORTER_OTLP_HEADERS", "k=v")},
		{name: "credential key in config", config: strs("OTEL_EXPORTER_OTLP_HEADERS", "k=v"), wantError: []string{"OTEL_EXPORTER_OTLP_HEADERS", "sensitive_config"}},
		{name: "proxy url in config", config: strs("DESTINATIONS_PROXY_URL", "http://u:p@proxy:8080"), wantError: []string{"DESTINATIONS_PROXY_URL", "sensitive_config"}},
		{name: "credential key in sensitive_config", sensitive: strs("PUBLISH_RABBITMQ_SERVER_URL", "amqp://u:p@h")},
		{name: "access key id in either map", config: strs("PUBLISH_AWS_SQS_ACCESS_KEY_ID", "AKIA")},
		{name: "access key id in sensitive_config", sensitive: strs("PUBLISH_AWS_SQS_ACCESS_KEY_ID", "AKIA")},
		{name: "key in both maps", config: strs("TOPICS", "a"), sensitive: strs("TOPICS", "a"), wantError: []string{"TOPICS", "both"}},
		{name: "operator events url in config", config: strs("OPERATOR_EVENTS_HTTP_URL", "x"), wantError: []string{"OPERATOR_EVENTS_HTTP_URL", "operator event destinations"}},
		{name: "operator events topics in sensitive_config", sensitive: strs("OPERATOR_EVENTS_TOPICS", "x"), wantError: []string{"OPERATOR_EVENTS_TOPICS"}},
		{name: "webhook proxy alias in config", config: strs("DESTINATIONS_WEBHOOK_PROXY_URL", "x"), wantError: []string{"DESTINATIONS_WEBHOOK_PROXY_URL", "DESTINATIONS_PROXY_URL"}},
		{name: "webhook proxy alias in sensitive_config", sensitive: strs("DESTINATIONS_WEBHOOK_PROXY_URL", "x"), wantError: []string{"DESTINATIONS_WEBHOOK_PROXY_URL"}},
		{name: "unknown key in config warns", config: strs("TOPIC", "a"), wantWarn: []string{"TOPIC", "Check the spelling"}},
		{name: "unknown key in sensitive_config warns", sensitive: strs("NEW_SECRET", "a"), wantWarn: []string{"NEW_SECRET"}},
		{name: "null value in config", config: map[string]types.String{"TOPICS": types.StringNull()}, wantError: []string{"TOPICS", "Remove the key"}},
		{name: "null value in sensitive_config", sensitive: map[string]types.String{"OTEL_EXPORTER_OTLP_HEADERS": types.StringNull()}, wantError: []string{"OTEL_EXPORTER_OTLP_HEADERS", "Remove the key"}},
		{name: "null and unknown maps", config: nil, sensitive: nil},
		{name: "unknown value", config: map[string]types.String{"TOPICS": types.StringUnknown()}},
		{name: "unknown value of a url key", config: map[string]types.String{"OTEL_EXPORTER_OTLP_ENDPOINT": types.StringUnknown()}},
		{name: "url without credentials in config", config: strs("OTEL_EXPORTER_OTLP_ENDPOINT", "https://otel.example.com:4318/v1?x=a@b")},
		{name: "url with user and password in config", config: strs("OTEL_EXPORTER_OTLP_ENDPOINT", "https://user:token@otel.example.com"), wantError: []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "sensitive_config"}},
		{name: "url with user only in config", config: strs("PORTAL_LOGO", "https://user@cdn.example.com/logo.png"), wantError: []string{"PORTAL_LOGO", "sensitive_config"}},
		{name: "url with credentials under an unknown key", config: strs("NEW_URL", "https://user:token@example.com"), wantError: []string{"NEW_URL", "sensitive_config"}},
		{name: "unparseable url with credentials in config", config: strs("PUBLISH_AWS_SQS_ENDPOINT", "https://user:p%ss@sqs.example.com"), wantError: []string{"PUBLISH_AWS_SQS_ENDPOINT", "sensitive_config"}},
		{name: "url with credentials in sensitive_config", sensitive: strs("OTEL_EXPORTER_OTLP_ENDPOINT", "https://user:token@otel.example.com")},
		{name: "email address is not a url", config: strs("PORTAL_ORGANIZATION_NAME", "support@example.com")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateConfig(tc.config, tc.sensitive)
			errs, warns := messages(diags, diag.SeverityError), messages(diags, diag.SeverityWarning)
			if (errs == "") != (len(tc.wantError) == 0) || (warns == "") != (len(tc.wantWarn) == 0) {
				t.Fatalf("errors = %q, warnings = %q; want errors %v, warnings %v", errs, warns, tc.wantError, tc.wantWarn)
			}
			for _, want := range tc.wantError {
				if !strings.Contains(errs, want) {
					t.Errorf("error %q does not contain %q", errs, want)
				}
			}
			for _, want := range tc.wantWarn {
				if !strings.Contains(warns, want) {
					t.Errorf("warning %q does not contain %q", warns, want)
				}
			}
		})
	}
}

func TestValidateConfig_knownKeysNeverWarn(t *testing.T) {
	config, sensitive := map[string]types.String{}, map[string]types.String{}
	for k := range knownKeys {
		if _, rejected := rejectedKeys[k]; rejected {
			continue
		}
		if credentialKeys[k] {
			sensitive[k] = types.StringValue("v")
		} else {
			config[k] = types.StringValue("v")
		}
	}
	if diags := validateConfig(config, sensitive); len(diags) != 0 {
		t.Errorf("diagnostics for known keys: %v", diags)
	}
	if len(knownKeys) != 76 {
		t.Errorf("known keys = %d, want the 76 ManagedConfig keys", len(knownKeys))
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
		{name: "empty proxy against null", desired: map[string]string{"DESTINATIONS_PROXY_URL": ""}, prior: nil, want: map[string]*string{}},
		{name: "empty string against a default is sent", desired: map[string]string{"TOPICS": ""}, prior: nil, want: map[string]*string{"TOPICS": ptr("")}},
		{name: "empty string against null is sent for other keys", desired: map[string]string{"PORTAL_LOGO": ""}, prior: nil, want: map[string]*string{"PORTAL_LOGO": ptr("")}},
		{name: "new key", desired: map[string]string{"OTEL_SERVICE_NAME": "svc"}, prior: nil, want: map[string]*string{"OTEL_SERVICE_NAME": ptr("svc")}},
		{name: "empty", desired: nil, prior: nil, want: map[string]*string{}},
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

func TestSplitConfig(t *testing.T) {
	live := map[string]*string{
		"TOPICS":                         ptr("a"),
		"MAX_RETRY_LIMIT":                ptr("10"),
		"PORTAL_LOGO":                    nil,
		"OTEL_EXPORTER_OTLP_HEADERS":     ptr("k=v"),
		"NEW_FUTURE_KEY":                 ptr("future"),
		"DESTINATIONS_WEBHOOK_PROXY_URL": ptr("http://u:p@proxy"),
		"DESTINATIONS_PROXY_URL":         ptr("http://u:p@proxy"),
	}
	plain, sensitive := splitConfig(live)

	if want := map[string]string{"TOPICS": "a", "MAX_RETRY_LIMIT": "10"}; !reflect.DeepEqual(plain, want) {
		t.Errorf("plain = %v, want %v", plain, want)
	}
	wantSensitive := map[string]string{
		"OTEL_EXPORTER_OTLP_HEADERS":     "k=v",
		"NEW_FUTURE_KEY":                 "future",
		"DESTINATIONS_WEBHOOK_PROXY_URL": "http://u:p@proxy",
		"DESTINATIONS_PROXY_URL":         "http://u:p@proxy",
	}
	if !reflect.DeepEqual(sensitive, wantSensitive) {
		t.Errorf("sensitive = %v, want %v", sensitive, wantSensitive)
	}
}

func TestSplitConfig_urlValues(t *testing.T) {
	tests := []struct {
		name, key, value string
		wantSensitive    bool
	}{
		{name: "url", key: "OTEL_EXPORTER_OTLP_ENDPOINT", value: "https://otel.example.com:4318/v1"},
		{name: "url with an at sign in the query", key: "PORTAL_REFERER_URL", value: "https://example.com/?next=a@b"},
		{name: "empty url key", key: "PORTAL_LOGO", value: ""},
		{name: "user and password", key: "OTEL_EXPORTER_OTLP_ENDPOINT", value: "https://user:token@otel.example.com", wantSensitive: true},
		{name: "user only", key: "PUBLISH_AWS_SQS_ENDPOINT", value: "http://user@localhost:4566", wantSensitive: true},
		{name: "surrounding space", key: "PORTAL_LOGO", value: " https://user:token@cdn.example.com/logo.png", wantSensitive: true},
		{name: "userinfo under a key that is not a url key", key: "HTTP_USER_AGENT", value: "https://user:token@example.com", wantSensitive: true},
		{name: "unparseable with userinfo under a key that is not a url key", key: "HTTP_USER_AGENT", value: "https://user:p%ss@example.com", wantSensitive: true},
		{name: "unparseable url key", key: "PORTAL_FAVICON_URL", value: "https://exa mple.com/%zz", wantSensitive: true},
		{name: "url key without a host", key: "OTEL_EXPORTER_OTLP_ENDPOINT", value: "user:token@otel.example.com:4318", wantSensitive: true},
		{name: "url key with backslashes", key: "PORTAL_LOGO_DARK", value: `https:\\user:token@cdn.example.com`, wantSensitive: true},
		{name: "plain text under a key that is not a url key", key: "PORTAL_ORGANIZATION_NAME", value: "support@example.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plain, sensitive := splitConfig(map[string]*string{tc.key: ptr(tc.value)})
			_, inPlain := plain[tc.key]
			got, inSensitive := sensitive[tc.key]
			if inPlain == tc.wantSensitive || inSensitive != tc.wantSensitive || (inSensitive && got != tc.value) {
				t.Errorf("plain = %v, sensitive = %v, want sensitive %v", plain, sensitive, tc.wantSensitive)
			}
		})
	}
}

func TestSplitConfig_urlKeysAreNonSensitive(t *testing.T) {
	for k := range urlKeys {
		if !nonSensitiveKeys[k] {
			t.Errorf("%s is a url key but not a non-sensitive key", k)
		}
	}
}

func TestSplitConfig_credentialKeysNeverPlain(t *testing.T) {
	live := map[string]*string{}
	for k := range credentialKeys {
		live[k] = ptr("secret")
	}
	plain, sensitive := splitConfig(live)
	if len(plain) != 0 || len(sensitive) != len(credentialKeys) {
		t.Errorf("plain = %v, sensitive = %v", plain, sensitive)
	}
}
