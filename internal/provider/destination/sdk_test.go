package destination

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"terraform-provider-hookdeck/internal/sdkclient"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDeliveryPolicyPayload(t *testing.T) {
	tests := []struct {
		name, prior, config, want string
		invalid                   bool
	}{
		{name: "new policy", config: `{"delivery_policy":{"rate":10,"period":"second","groups":{"key":"body.id","rate":2,"rate_period":"second"}}}`, want: `{"delivery_policy":{"rate":10,"period":"second","groups":{"key":"body.id","rate":2,"rate_period":"second"}}}`},
		{name: "legacy rates", config: `{"rate_limit":10,"rate_limit_period":"concurrent"}`, want: `{"delivery_policy":{"rate":10,"period":"concurrent"}}`},
		{name: "legacy groups", config: `{"delivery_groups":{"key":"body.id","rate_limit":2,"rate_limit_period":"second","overrides":{"a":{"rate_limit":3,"rate_limit_period":"minute"}}}}`, want: `{"delivery_policy":{"groups":{"key":"body.id","rate":2,"rate_period":"second","overrides":{"a":{"rate":3,"rate_period":"minute"}}}}}`},
		{name: "mixed rejected", config: `{"delivery_policy":{},"rate_limit":2}`, invalid: true},
		{name: "null policy mixed rejected", config: `{"delivery_policy":null,"delivery_groups":null}`, invalid: true},
		{name: "legacy null", config: `{"rate_limit":null}`, want: `{"delivery_policy":{"rate":null}}`},
		{name: "legacy to new", prior: `{"rate_limit":10,"rate_limit_period":"second"}`, config: `{"delivery_policy":{"rate":10,"period":"second"}}`, want: `{"delivery_policy":{"rate":10,"period":"second"}}`},
		{name: "new to legacy", prior: `{"delivery_policy":{"rate":10,"period":"second"}}`, config: `{"rate_limit":10,"rate_limit_period":"second"}`, want: `{"delivery_policy":{"rate":10,"period":"second"}}`},
		{name: "remove legacy rate", prior: `{"rate_limit":10,"rate_limit_period":"second"}`, config: `{}`, want: `{"delivery_policy":null}`},
		{name: "remove whole policy", prior: `{"delivery_policy":{"rate":10,"groups":{"key":"body.id"}}}`, config: `{}`, want: `{"delivery_policy":null}`},
		{name: "remove group and period", prior: `{"delivery_policy":{"rate":10,"period":"minute","groups":{"key":"body.id"}}}`, config: `{"delivery_policy":{"rate":5}}`, want: `{"delivery_policy":{"rate":5,"period":null,"groups":null}}`},
		{name: "remove rate retain period", prior: `{"rate_limit":10,"rate_limit_period":"concurrent"}`, config: `{"rate_limit_period":"concurrent"}`, want: `{"delivery_policy":{"rate":null,"period":"concurrent"}}`},
		{name: "replace overrides", prior: `{"delivery_policy":{"groups":{"key":"body.id","overrides":{"a":{"rate":3,"rate_period":"second"}}}}}`, config: `{"delivery_policy":{"groups":{"key":"body.id"}}}`, want: `{"delivery_policy":{"groups":{"key":"body.id"}}}`},
		{name: "explicit null", prior: `{"delivery_policy":{"rate":10}}`, config: `{"delivery_policy":null}`, want: `{"delivery_policy":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := destinationResourceModel{Config: jsontypes.NewNormalizedValue(tt.config), Type: types.StringValue("HTTP")}
			var prior *destinationResourceModel
			if tt.prior != "" {
				prior = &destinationResourceModel{Config: jsontypes.NewNormalizedValue(tt.prior)}
			}
			payload, diags := m.toUpdatePayload(prior)
			if tt.invalid {
				if !diags.HasError() {
					t.Fatal("expected error")
				}
				return
			}
			if diags.HasError() {
				t.Fatal(diags)
			}
			var want map[string]interface{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(payload["config"], want) {
				t.Fatalf("got %#v; want %#v", payload["config"], want)
			}
			if m.Config.ValueString() != tt.config {
				t.Fatal("request conversion changed Terraform config")
			}
		})
	}
}

func TestDestinationUsesNewAPIVersion(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		expected := "/2026-09-01/destinations"
		if r.Method != "POST" {
			expected += "/des_test"
		}
		if r.URL.Path != expected {
			t.Errorf("path = %s, want %s", r.URL.Path, expected)
		}
		if r.Method == "POST" || r.Method == "PUT" {
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			config, ok := payload["config"].(map[string]interface{})
			if !ok {
				t.Error("missing config object")
				return
			}
			if _, ok := config["rate_limit"]; ok {
				t.Error("sent legacy field to new API")
			}
			if _, ok := config["delivery_policy"]; !ok {
				t.Error("missing delivery policy")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"des_test","name":"test","team_id":"tm_test","created_at":"2026-09-01T00:00:00Z","updated_at":"2026-09-01T00:00:00Z","type":"HTTP"}`)
	}))
	defer server.Close()
	client := sdkclient.InitHookdeckSDKClient(server.URL, "test", "test")
	m := destinationResourceModel{Name: types.StringValue("test"), Config: jsontypes.NewNormalizedValue(`{"rate_limit":10}`)}
	ctx := t.Context()
	if d := m.Create(ctx, &client); d.HasError() {
		t.Fatal(d)
	}
	if d := m.Update(ctx, &client, nil); d.HasError() {
		t.Fatal(d)
	}
	if d := m.Retrieve(ctx, &client); d.HasError() {
		t.Fatal(d)
	}
	if d := m.Delete(ctx, &client); d.HasError() {
		t.Fatal(d)
	}
	if !reflect.DeepEqual(methods, []string{"POST", "PUT", "GET", "DELETE"}) {
		t.Fatal(methods)
	}
}
