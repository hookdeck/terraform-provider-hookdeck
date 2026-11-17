package shared

import (
	"context"
	"testing"

	"terraform-provider-hookdeck/internal/configjson"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPlanConfig(t *testing.T) {
	full := `{"url":"https://x","http_method":null,"auth_type":"HOOKDECK_SIGNATURE","auth":{},"delivery_policy":{"rate":5,"period":"second"}}`
	str := types.StringValue
	null := types.StringNull()
	unknown := types.StringUnknown()
	managed := configjson.Paths{"url": {}, "delivery_policy": {"rate": {}}}

	tests := []struct {
		name          string
		config, state types.String
		plan          types.String
		managed       configjson.Paths
		recorded      bool
		want          types.String
		wantMissing   []string
	}{
		{name: "create keeps the proposed value", config: str(`{"url":"https://x"}`), state: null, plan: str(`{"url":"https://x"}`), want: str(`{"url":"https://x"}`)},
		{name: "create without config stays unknown", config: null, state: null, plan: unknown, want: unknown},
		{name: "config covers state", config: str(`{"url":"https://x","delivery_policy":{"rate":5}}`), state: str(full), plan: str(`{"url":"https://x","delivery_policy":{"rate":5}}`), managed: managed, recorded: true, want: str(full)},
		{name: "config differs", config: str(`{"url":"https://y"}`), state: str(full), plan: str(`{"url":"https://y"}`), managed: managed, recorded: true, want: str(`{"url":"https://y"}`)},
		{name: "managed key removed while the rest covers", config: str(`{"url":"https://x"}`), state: str(full), plan: str(`{"url":"https://x"}`), managed: managed, recorded: true, want: str(`{"url":"https://x"}`)},
		{name: "no config and no managed keys", config: null, state: str(full), plan: str(full), managed: configjson.Paths{}, recorded: true, want: str(full)},
		{name: "no config and managed keys", config: null, state: str(full), plan: str(full), managed: managed, recorded: true, want: unknown},
		{name: "config unknown", config: unknown, state: str(full), plan: unknown, managed: managed, recorded: true, want: unknown},
		{name: "managed key Hookdeck does not return", config: str(`{"url":"https://x","typo":1}`), state: str(`{"url":"https://x"}`), plan: str(`{"url":"https://x","typo":1}`), managed: configjson.Paths{"url": {}, "typo": {}}, recorded: true, want: str(`{"url":"https://x"}`), wantMissing: []string{"typo"}},
		{name: "new key Hookdeck does not return", config: str(`{"url":"https://x","typo":1}`), state: str(`{"url":"https://x"}`), plan: str(`{"url":"https://x","typo":1}`), managed: configjson.Paths{"url": {}}, recorded: true, want: str(`{"url":"https://x","typo":1}`)},
		{name: "imported and never applied", config: str(`{"url":"https://x"}`), state: str(full), plan: str(`{"url":"https://x"}`), want: str(full)},
		{name: "imported and never applied, no config", config: null, state: str(full), plan: str(full), want: str(full)},
		{name: "state holds the applied config and a key is added", config: str(`{"url":"https://x","http_method":"POST"}`), state: str(`{"url":"https://x"}`), plan: str(`{"url":"https://x","http_method":"POST"}`), managed: configjson.Paths{"url": {}}, recorded: true, want: str(`{"url":"https://x","http_method":"POST"}`)},
		{name: "whitespace and key order are not a change", config: str("{\n \"delivery_policy\": {\"rate\": 5},\n \"url\": \"https://x\"\n}"), state: str(full), plan: str("{}"), managed: managed, recorded: true, want: str(full)},
		{name: "state is not JSON", config: str(`{"url":"https://x"}`), state: str(`oops`), plan: str(`{"url":"https://x"}`), want: str(`{"url":"https://x"}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, missing, err := planConfig(tt.config, tt.state, tt.plan, tt.managed, tt.recorded)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("planned %s, want %s", got, tt.want)
			}
			if len(missing) != len(tt.wantMissing) || (len(missing) > 0 && missing[0] != tt.wantMissing[0]) {
				t.Fatalf("missing = %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func TestPlanConfigInvalidConfig(t *testing.T) {
	for _, config := range []string{`[1]`, `{`} {
		if _, _, err := planConfig(types.StringValue(config), types.StringValue(`{}`), types.StringValue(config), nil, true); err == nil {
			t.Errorf("%s: expected an error", config)
		}
	}
}

type fakePrivate map[string][]byte

func (f fakePrivate) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	return f[key], nil
}

func (f fakePrivate) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	f[key] = value
	return nil
}

func TestManagedConfigPaths(t *testing.T) {
	ctx := context.Background()
	p := fakePrivate{}

	if _, recorded, _ := ManagedConfigPaths(ctx, p); recorded {
		t.Fatal("nothing recorded yet")
	}
	if d := SetManagedConfigPaths(ctx, p, jsontypes.NewNormalizedValue(`{"url":"x","delivery_policy":{"rate":1}}`)); d.HasError() {
		t.Fatal(d)
	}
	if string(p[configKeysPrivateKey]) != `{"v":1,"keys":{"delivery_policy":{"rate":{}},"url":{}}}` {
		t.Fatalf("stored %s", p[configKeysPrivateKey])
	}
	paths, recorded, d := ManagedConfigPaths(ctx, p)
	if d.HasError() || !recorded || len(paths) != 2 || len(paths["delivery_policy"]) != 1 {
		t.Fatalf("got %v %v %v", paths, recorded, d)
	}

	if d := SetManagedConfigPaths(ctx, p, jsontypes.NewNormalizedNull()); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, p); !recorded || len(paths) != 0 {
		t.Fatalf("null config records an empty set, got %v %v", paths, recorded)
	}

	p[configKeysPrivateKey] = []byte(`{"v":9,"keys":{}}`)
	if _, recorded, d := ManagedConfigPaths(ctx, p); recorded || d.HasError() {
		t.Fatalf("unknown version is treated as not recorded: %v %v", recorded, d)
	}
}

func TestSeedManagedConfigPaths(t *testing.T) {
	ctx := context.Background()

	p := fakePrivate{}
	if d := SeedManagedConfigPaths(ctx, p, jsontypes.NewNormalizedValue(`{"url":"x"}`)); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, p); !recorded || len(paths) != 1 {
		t.Fatalf("seeded from state config: %v %v", paths, recorded)
	}

	// Already recorded: state is the full config and must not overwrite the set.
	if d := SeedManagedConfigPaths(ctx, p, jsontypes.NewNormalizedValue(`{"url":"x","auth":{}}`)); d.HasError() {
		t.Fatal(d)
	}
	if paths, _, _ := ManagedConfigPaths(ctx, p); len(paths) != 1 {
		t.Fatalf("recorded set was replaced: %v", paths)
	}

	imported := fakePrivate{}
	if d := SeedManagedConfigPaths(ctx, imported, jsontypes.NewNormalizedNull()); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, imported); !recorded || len(paths) != 0 {
		t.Fatalf("import seeds an empty set: %v %v", paths, recorded)
	}
}

func TestConfigFromAPI(t *testing.T) {
	obj, err := DecodeObject([]byte(`{"id":"x","config":{"b":1152921504606846976,"a":null}}`))
	if err != nil {
		t.Fatal(err)
	}
	got, diags := ConfigFromAPI(obj)
	if diags.HasError() || got.ValueString() != `{"a":null,"b":1152921504606846976}` {
		t.Fatalf("got %s, %v", got.ValueString(), diags)
	}
	for _, body := range []string{`{"config":null}`, `{}`} {
		obj, _ := DecodeObject([]byte(body))
		got, diags := ConfigFromAPI(obj)
		if diags.HasError() || got.ValueString() != `{}` {
			t.Fatalf("%s: got %s, %v", body, got.ValueString(), diags)
		}
	}
	if _, err := DecodeObject([]byte(`<html>`)); err == nil {
		t.Fatal("expected an error")
	}
}
