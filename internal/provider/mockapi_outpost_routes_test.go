package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// mockOutpost is the configuration of one Outpost project in the mock
// Outpost API.
type mockOutpost struct {
	overrides map[string]string
	// notProvisioned is the number of reads still answered with 409; a
	// negative value is never provisioned.
	notProvisioned int
	// lockedPatches is the number of PATCH requests still answered with 409
	// OUTPOST_CONFIG_LOCKED.
	lockedPatches int
}

// mockOutpostDefaults are the values the API returns for a key nobody set.
var mockOutpostDefaults = map[string]string{
	"MAX_RETRY_LIMIT":           "10",
	"IDGEN_TYPE":                "nanoid",
	"DESTINATIONS_WEBHOOK_MODE": "default",
}

// mockOutpostKeys are the keys the mock API accepts. EXTRA_FUTURE_KEY is
// one the provider does not know.
var mockOutpostKeys = []string{
	"TOPICS", "RETRY_INTERVAL_SECONDS", "MAX_RETRY_LIMIT", "IDGEN_TYPE", "DESTINATIONS_WEBHOOK_MODE",
	"OTEL_SERVICE_NAME", "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_HEADERS",
	"DESTINATIONS_PROXY_URL", "DESTINATIONS_WEBHOOK_PROXY_URL", "PORTAL_LOGO",
	"PUBLISH_RABBITMQ_SERVER_URL", "EXTRA_FUTURE_KEY",
}

func isMockOutpostKey(key string) bool {
	for _, k := range mockOutpostKeys {
		if k == key {
			return true
		}
	}
	return false
}

func (o *mockOutpost) view() map[string]*string {
	out := map[string]*string{}
	for _, k := range mockOutpostKeys {
		if v, ok := o.overrides[k]; ok {
			value := v
			out[k] = &value
		} else if v, ok := mockOutpostDefaults[k]; ok {
			value := v
			out[k] = &value
		} else {
			out[k] = nil
		}
	}
	// The proxy URL and its alias mirror each other.
	out["DESTINATIONS_WEBHOOK_PROXY_URL"] = out["DESTINATIONS_PROXY_URL"]
	return out
}

func (m *mockAPI) handleOutpost(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := apiVersionPrefix.ReplaceAllString(r.URL.Path, "")
	apiKey := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	header := r.Header.Get("x-hookdeck-team-id")
	m.requests = append(m.requests, mockRequest{key: apiKey, method: r.Method, path: path, project: header})

	key, ok := m.keys[apiKey]
	if !ok {
		unauthorized(w)
		return
	}
	projectID, ok := m.project(w, key, header)
	if !ok {
		return
	}
	outpost, ok := m.outposts[projectID]
	if !ok || path != "/config" {
		apiError(w, 404, "NOT_FOUND", "Not Found", nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if outpost.notProvisioned != 0 {
			if outpost.notProvisioned > 0 {
				outpost.notProvisioned--
			}
			apiError(w, 409, "OUTPOST_NOT_PROVISIONED", "Outpost is not provisioned for this project", nil)
			return
		}
		writeJSON(w, 200, outpost.view())
	case http.MethodPatch:
		m.patchOutpost(w, r, outpost)
	default:
		apiError(w, 404, "NOT_FOUND", "Not Found", nil)
	}
}

func (m *mockAPI) patchOutpost(w http.ResponseWriter, r *http.Request, outpost *mockOutpost) {
	var patch map[string]*string
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		apiError(w, 422, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	m.patchBodies = append(m.patchBodies, patch)
	if outpost.notProvisioned != 0 {
		apiError(w, 409, "OUTPOST_NOT_PROVISIONED", "Outpost is not provisioned for this project", nil)
		return
	}
	if outpost.lockedPatches > 0 {
		outpost.lockedPatches--
		apiError(w, 409, "OUTPOST_CONFIG_LOCKED", "The configuration is being updated", nil)
		return
	}
	for k := range patch {
		if k == "OPERATOR_EVENTS_HTTP_URL" || k == "OPERATOR_EVENTS_TOPICS" || !isMockOutpostKey(k) {
			apiError(w, 422, "VALIDATION_ERROR", fmt.Sprintf("%q is not allowed", k), nil)
			return
		}
	}
	for k, v := range patch {
		if k == "DESTINATIONS_WEBHOOK_PROXY_URL" {
			k = "DESTINATIONS_PROXY_URL"
		}
		switch {
		case v == nil:
			if def, ok := mockOutpostDefaults[k]; ok {
				outpost.overrides[k] = def
			} else {
				delete(outpost.overrides, k)
			}
		case k == "DESTINATIONS_PROXY_URL" && *v == "":
			delete(outpost.overrides, k)
		default:
			outpost.overrides[k] = *v
		}
	}
	writeJSON(w, 200, outpost.view())
}

// sentPatches returns the bodies of the PATCH requests so far.
func (m *mockAPI) sentPatches() []map[string]*string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]map[string]*string(nil), m.patchBodies...)
}

const (
	outpostAddr        = "hookdeck_outpost_project.test"
	outpostDataAddr    = "data.hookdeck_outpost_project.test"
	mockOutpostKeyName = "hd_mock_outpost"
)

func outpostConfig(name, body string) string {
	return fmt.Sprintf("resource \"hookdeck_outpost_project\" \"test\" {\n  name = %q\n%s}\n", name, body)
}

func ptr(s string) *string { return &s }

func countRequests(m *mockAPI, method, prefix string) int {
	n := 0
	for _, r := range m.requestsTo(prefix) {
		if r.method == method {
			n++
		}
	}
	return n
}

func checkNoNewWrites(m *mockAPI, patches, puts *int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := countRequests(m, http.MethodPatch, "/config") - *patches; got != 0 {
			return fmt.Errorf("%d PATCH /config requests, want none", got)
		}
		if got := countRequests(m, http.MethodPut, "/projects") - *puts; got != 0 {
			return fmt.Errorf("%d PUT /projects requests, want none", got)
		}
		return nil
	}
}

func checkSentPatches(m *mockAPI, want ...map[string]*string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := m.sentPatches(); !reflect.DeepEqual(got, want) {
			return fmt.Errorf("PATCH bodies = %s, want %s", formatPatches(got), formatPatches(want))
		}
		return nil
	}
}

func formatPatches(patches []map[string]*string) string {
	out := "["
	for _, p := range patches {
		out += "{"
		for k, v := range p {
			if v == nil {
				out += k + ":null "
			} else {
				out += fmt.Sprintf("%s:%q ", k, *v)
			}
		}
		out += "}"
	}
	return out + "]"
}

func replacePlan() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(outpostAddr, plancheck.ResourceActionReplace),
	}}
}

// mockOutpostOrganization is mockOrganization with the organization key as
// the provider's key.
func mockOutpostOrganization(t *testing.T) *mockAPI {
	t.Helper()
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockOrgKey)
	return m
}

const outpostBasic = `  config = {
    TOPICS                 = "a,b"
    RETRY_INTERVAL_SECONDS = "60"
  }
  sensitive_config = {
    OTEL_EXPORTER_OTLP_HEADERS = "k=v"
  }
`
