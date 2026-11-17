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
		{name: "key added that Hookdeck already holds", config: str(`{"url":"https://x","delivery_policy":{"rate":5}}`), state: str(full), plan: str(`{"url":"https://x","delivery_policy":{"rate":5}}`), managed: configjson.Paths{"url": {}}, recorded: true, want: str(`{"url":"https://x","delivery_policy":{"rate":5}}`)},
		{name: "nested key added that Hookdeck already holds", config: str(`{"url":"https://x","delivery_policy":{"rate":5,"period":"second"}}`), state: str(full), plan: str(`{"url":"https://x","delivery_policy":{"rate":5,"period":"second"}}`), managed: managed, recorded: true, want: str(`{"url":"https://x","delivery_policy":{"rate":5,"period":"second"}}`)},
		{name: "recorded with no keys and a key added", config: str(`{"url":"https://x"}`), state: str(full), plan: str(`{"url":"https://x"}`), managed: configjson.Paths{}, recorded: true, want: str(`{"url":"https://x"}`)},
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
	none := types.StringNull()

	if _, recorded, _ := ManagedConfigPaths(ctx, p, types.StringUnknown()); recorded {
		t.Fatal("nothing recorded yet")
	}
	if d := SetManagedConfigPaths(ctx, p, jsontypes.NewNormalizedValue(`{"url":"x","delivery_policy":{"rate":1}}`)); d.HasError() {
		t.Fatal(d)
	}
	if string(p[configKeysPrivateKey]) != `{"v":1,"keys":{"delivery_policy":{"rate":{}},"url":{}}}` {
		t.Fatalf("stored %s", p[configKeysPrivateKey])
	}
	// State is the full config: stored paths win over it.
	paths, recorded, d := ManagedConfigPaths(ctx, p, types.StringValue(`{"url":"x","auth":{}}`))
	if d.HasError() || !recorded || len(paths) != 2 || len(paths["delivery_policy"]) != 1 {
		t.Fatalf("got %v %v %v", paths, recorded, d)
	}

	if d := SetManagedConfigPaths(ctx, p, jsontypes.NewNormalizedNull()); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, p, none); !recorded || len(paths) != 0 {
		t.Fatalf("null config records an empty set, got %v %v", paths, recorded)
	}
}

// Private data that is missing derives the keys from state, which in that case
// holds exactly the user's config. Private data that exists but cannot be read
// is never derived from state, which holds Hookdeck's full config.
func TestManagedConfigPathsFromState(t *testing.T) {
	ctx := context.Background()
	state := types.StringValue(`{"url":"x","auth":{"token":"t"}}`)

	paths, recorded, d := ManagedConfigPaths(ctx, fakePrivate{}, state)
	if d.HasError() || !recorded || len(paths) != 2 || len(paths["auth"]) != 1 {
		t.Fatalf("absent private data: %v %v %v", paths, recorded, d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, fakePrivate{}, types.StringNull()); !recorded || len(paths) != 0 {
		t.Fatal("absent private data and no config in state records an empty set")
	}
	if _, recorded, _ := ManagedConfigPaths(ctx, fakePrivate{}, types.StringValue(`oops`)); recorded {
		t.Fatal("absent private data and unreadable state is not recorded")
	}

	for name, raw := range map[string]string{
		"unknown version": `{"v":9,"keys":{}}`,
		"not json":        `oops`,
		"no version":      `{"keys":{}}`,
	} {
		p := fakePrivate{configKeysPrivateKey: []byte(raw)}
		if paths, recorded, d := ManagedConfigPaths(ctx, p, state); recorded || len(paths) != 0 || d.HasError() {
			t.Fatalf("%s: got %v %v %v, want not recorded", name, paths, recorded, d)
		}
	}
}

func TestSeedManagedConfigPaths(t *testing.T) {
	ctx := context.Background()
	full := jsontypes.NewNormalizedValue(`{"url":"x","auth":{}}`)

	// State from a version that did not track the keys: its config is the user's.
	p := fakePrivate{}
	if d := SeedManagedConfigPaths(ctx, p, jsontypes.NewNormalizedValue(`{"url":"x"}`)); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, p, types.StringNull()); !recorded || len(paths) != 1 {
		t.Fatalf("seeded from state config: %v %v", paths, recorded)
	}

	// Already recorded: state is the full config and must not overwrite the set.
	if d := SeedManagedConfigPaths(ctx, p, full); d.HasError() {
		t.Fatal(d)
	}
	if paths, _, _ := ManagedConfigPaths(ctx, p, types.StringNull()); len(paths) != 1 {
		t.Fatalf("recorded set was replaced: %v", paths)
	}

	// No private data and no config in state (created without config): an
	// empty set is recorded, so a key added later counts as new.
	bare := fakePrivate{}
	if d := SeedManagedConfigPaths(ctx, bare, jsontypes.NewNormalizedNull()); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, bare, full.StringValue); !recorded || len(paths) != 0 {
		t.Fatalf("created without config: %v %v", paths, recorded)
	}

	// Import: ImportState stores the marker; a later read, when state holds
	// the full config, does not seed from it.
	marker, err := configjson.MarshalUnrecorded()
	if err != nil {
		t.Fatal(err)
	}
	imported := fakePrivate{configKeysPrivateKey: marker}
	if paths, recorded, _ := ManagedConfigPaths(ctx, imported, full.StringValue); recorded || len(paths) != 0 {
		t.Fatalf("import is not recorded: %v %v", paths, recorded)
	}
	if d := SeedManagedConfigPaths(ctx, imported, full); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, imported, full.StringValue); recorded || len(paths) != 0 {
		t.Fatalf("later read seeded from the full config: %v %v", paths, recorded)
	}

	// Unreadable private data is left alone, not reseeded from state.
	broken := fakePrivate{configKeysPrivateKey: []byte(`{"v":9,"keys":{}}`)}
	if d := SeedManagedConfigPaths(ctx, broken, full); d.HasError() {
		t.Fatal(d)
	}
	if string(broken[configKeysPrivateKey]) != `{"v":9,"keys":{}}` {
		t.Fatalf("unreadable data was replaced: %s", broken[configKeysPrivateKey])
	}
	if _, recorded, _ := ManagedConfigPaths(ctx, broken, full.StringValue); recorded {
		t.Fatal("unreadable data counts as not recorded")
	}

	// The first apply records the keys.
	if d := SetManagedConfigPaths(ctx, imported, jsontypes.NewNormalizedValue(`{"url":"x"}`)); d.HasError() {
		t.Fatal(d)
	}
	if paths, recorded, _ := ManagedConfigPaths(ctx, imported, full.StringValue); !recorded || len(paths) != 1 {
		t.Fatalf("after the first apply: %v %v", paths, recorded)
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
