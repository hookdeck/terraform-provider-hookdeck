package source

import (
	"encoding/json"
	"reflect"
	"testing"

	"terraform-provider-hookdeck/internal/configjson"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSourceUpdatePayload(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		managed    string
		prior      string
		config     string
		want       string
	}{
		{name: "nothing managed", sourceType: "WEBHOOK", config: `{"a":1}`, want: `{"a":1}`},
		{name: "managed key kept", sourceType: "WEBHOOK", managed: `{"a":{}}`, prior: `{"a":1}`, config: `{"a":2}`, want: `{"a":2}`},
		{name: "removed key is cleared", sourceType: "WEBHOOK", managed: `{"a":{},"b":{}}`, prior: `{"a":1,"b":"x"}`, config: `{"a":1}`, want: `{"a":1,"b":null}`},
		{name: "keys Hookdeck added are not cleared", sourceType: "WEBHOOK", managed: `{"a":{}}`, prior: `{"a":1,"allowed_http_methods":["POST"],"auth":{"token":"t"}}`, config: `{"a":1}`, want: `{"a":1}`},
		{name: "array without a default is left out", sourceType: "WEBHOOK", managed: `{"a":{}}`, prior: `{"a":["x"]}`, config: `{}`, want: `{}`},
		{name: "removed array resets to the default", sourceType: "HTTP", managed: `{"allowed_http_methods":{}}`, prior: `{"allowed_http_methods":["GET"]}`, config: `{}`, want: `{"allowed_http_methods":["PUT","POST","PATCH","DELETE"]}`},
		{name: "removed boolean without a default is left out", sourceType: "HTTP", managed: `{"flag":{}}`, prior: `{"flag":true}`, config: `{}`, want: `{}`},
		{name: "removed key that Hookdeck did not return", sourceType: "WEBHOOK", managed: `{"typo":{}}`, prior: `{}`, config: `{}`, want: `{"typo":null}`},
		{name: "explicit null is sent as it is", sourceType: "WEBHOOK", managed: `{"a":{}}`, prior: `{"a":1}`, config: `{"a":null}`, want: `{"a":null}`},
		{name: "whole config removed", sourceType: "WEBHOOK", managed: `{"a":{},"b":{}}`, prior: `{"a":1,"b":2}`, want: `{"a":null,"b":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := sourceResourceModel{Type: types.StringValue(tt.sourceType), Name: types.StringValue("s")}
			userConfig := jsontypes.NewNormalizedNull()
			if tt.config != "" {
				userConfig = jsontypes.NewNormalizedValue(tt.config)
			}
			var managed configjson.Paths
			if tt.managed != "" {
				if err := json.Unmarshal([]byte(tt.managed), &managed); err != nil {
					t.Fatal(err)
				}
			}
			var prior *sourceResourceModel
			if tt.prior != "" {
				prior = &sourceResourceModel{Config: jsontypes.NewNormalizedValue(tt.prior)}
			}
			payload, diags := m.toUpdatePayload(userConfig, managed, prior)
			if diags.HasError() {
				t.Fatal(diags)
			}
			var want map[string]interface{}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			got, _ := payload["config"].(map[string]interface{})
			if got == nil {
				got = map[string]interface{}{}
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Fatalf("got %s; want %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestSourcePayloadSendsOnlyUserConfig(t *testing.T) {
	m := sourceResourceModel{
		Name:   types.StringValue("s"),
		Config: jsontypes.NewNormalizedValue(`{"auth":{"token":"t"},"allowed_http_methods":["POST"]}`),
	}
	payload, diags := m.toPayload(jsontypes.NewNormalizedValue(`{"allowed_http_methods":["POST"]}`))
	if diags.HasError() {
		t.Fatal(diags)
	}
	config, ok := payload["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("payload has no config: %v", payload)
	}
	if _, ok := config["auth"]; ok {
		t.Fatal("auth from the planned value was sent")
	}
	if _, ok := toPayloadNoConfig(t)["config"]; ok {
		t.Fatal("config sent although the user set none")
	}
}

func toPayloadNoConfig(t *testing.T) map[string]interface{} {
	t.Helper()
	m := sourceResourceModel{Name: types.StringValue("s"), Config: jsontypes.NewNormalizedUnknown()}
	payload, diags := m.toPayload(jsontypes.NewNormalizedNull())
	if diags.HasError() {
		t.Fatal(diags)
	}
	return payload
}
