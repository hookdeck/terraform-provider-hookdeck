package configjson

import (
	"reflect"
	"strings"
	"testing"
)

func mustParse(t *testing.T, s string) map[string]any {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("parse %s: %v", s, err)
	}
	return v
}

func TestCovers(t *testing.T) {
	tests := []struct {
		name        string
		user, state string
		managed     Paths
		want        bool
		wantMissing []string
	}{
		{name: "same keys in another order", user: `{"a": 1,  "b": "x"}`, state: `{"b":"x","a":1}`, want: true},
		{name: "state has extra keys", user: `{"url":"https://x"}`, state: `{"url":"https://x","http_method":null,"auth_type":"HOOKDECK_SIGNATURE","auth":{}}`, want: true},
		{name: "top level value differs", user: `{"url":"https://x"}`, state: `{"url":"https://y"}`, want: false},
		{name: "nested subset", user: `{"delivery_policy":{"rate":5}}`, state: `{"delivery_policy":{"rate":5,"period":"second","groups":null}}`, want: true},
		{name: "nested value differs", user: `{"delivery_policy":{"rate":5}}`, state: `{"delivery_policy":{"rate":6,"period":"second"}}`, want: false},
		{name: "nested object missing in state", user: `{"delivery_policy":{"rate":5}}`, state: `{}`, want: false},
		{name: "nested object is a scalar in state", user: `{"delivery_policy":{"rate":5}}`, state: `{"delivery_policy":"x"}`, want: false},
		{name: "array equal", user: `{"m":["PUT","POST"]}`, state: `{"m":["PUT","POST"]}`, want: true},
		{name: "array reordered", user: `{"m":["PUT","POST"]}`, state: `{"m":["POST","PUT"]}`, want: false},
		{name: "array shorter in state", user: `{"m":["PUT","POST"]}`, state: `{"m":["PUT"]}`, want: false},
		{name: "array longer in state", user: `{"m":["PUT"]}`, state: `{"m":["PUT","POST"]}`, want: false},
		{name: "array of objects with extra keys", user: `{"g":[{"key":"a"},{"key":"b"}]}`, state: `{"g":[{"key":"a","rate":1},{"key":"b","rate":2}]}`, want: true},
		{name: "array of objects value differs", user: `{"g":[{"key":"a"}]}`, state: `{"g":[{"key":"b"}]}`, want: false},
		{name: "user null and state absent", user: `{"x":null}`, state: `{}`, want: true},
		{name: "user null and state null", user: `{"x":null}`, state: `{"x":null}`, want: true},
		{name: "user null and state set", user: `{"x":null}`, state: `{"x":1}`, want: false},
		{name: "managed key Hookdeck dropped", user: `{"typo":"v"}`, state: `{}`, managed: Paths{"typo": {}}, want: true, wantMissing: []string{"typo"}},
		{name: "unmanaged key Hookdeck lacks", user: `{"typo":"v"}`, state: `{}`, want: false},
		{name: "managed nested key dropped", user: `{"p":{"a":1,"b":2}}`, state: `{"p":{"a":1}}`, managed: Paths{"p": {"a": {}, "b": {}}}, want: true, wantMissing: []string{"p.b"}},
		{name: "managed parent but new nested key", user: `{"p":{"a":1,"b":2}}`, state: `{"p":{"a":1}}`, managed: Paths{"p": {"a": {}}}, want: false},
		{name: "no keys", user: `{}`, state: `{"a":1}`, want: true},
		{name: "number written with a decimal point", user: `{"n":10}`, state: `{"n":10.0}`, want: true},
		{name: "exponent", user: `{"n":1e2}`, state: `{"n":100}`, want: true},
		{name: "large integer exact", user: `{"n":1152921504606846976}`, state: `{"n":1152921504606846976}`, want: true},
		{name: "large integers that differ by one", user: `{"n":1152921504606846976}`, state: `{"n":1152921504606846977}`, want: false},
		{name: "decimals", user: `{"n":0.1}`, state: `{"n":0.10}`, want: true},
		{name: "numeric string vs number", user: `{"n":"10"}`, state: `{"n":10}`, want: true},
		{name: "number vs numeric string", user: `{"n":10}`, state: `{"n":"10"}`, want: true},
		{name: "different numeric string", user: `{"n":"11"}`, state: `{"n":10}`, want: false},
		{name: "boolean string vs boolean", user: `{"b":"true"}`, state: `{"b":true}`, want: true},
		{name: "false string vs true", user: `{"b":"false"}`, state: `{"b":true}`, want: false},
		{name: "non numeric string vs number", user: `{"n":"abc"}`, state: `{"n":0}`, want: false},
		{name: "fraction string is not a number", user: `{"n":"1/2"}`, state: `{"n":0.5}`, want: false},
		{name: "hex string is not a number", user: `{"n":"0x10"}`, state: `{"n":16}`, want: false},
		{name: "huge exponent is not compared", user: `{"n":"1e999999999"}`, state: `{"n":1}`, want: false},
		{name: "strings compare exactly", user: `{"s":"Abc"}`, state: `{"s":"abc"}`, want: false},
		{name: "string vs boolean", user: `{"s":"yes"}`, state: `{"s":true}`, want: false},
		{name: "empty string vs null", user: `{"s":""}`, state: `{"s":null}`, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, missing := Covers(mustParse(t, tt.user), mustParse(t, tt.state), tt.managed)
			if got != tt.want {
				t.Fatalf("covers = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(missing, tt.wantMissing) {
				t.Fatalf("missing = %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func TestPathsOf(t *testing.T) {
	got := PathsOf(mustParse(t, `{"url":"x","delivery_policy":{"rate":1,"groups":{"key":"k"}},"allowed":["a",{"b":1}],"nothing":null}`))
	want := Paths{
		"url":             {},
		"delivery_policy": {"rate": {}, "groups": {"key": {}}},
		"allowed":         {},
		"nothing":         {},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestPathsRoundTrip(t *testing.T) {
	paths := Paths{"url": {}, "delivery_policy": {"rate": {}}}
	b, err := paths.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"v":1,"keys":{"delivery_policy":{"rate":{}},"url":{}}}` {
		t.Fatalf("envelope = %s", b)
	}
	got, err := UnmarshalPaths(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, paths) {
		t.Fatalf("got %v, want %v", got, paths)
	}

	empty, err := Paths(nil).Marshal()
	if err != nil || string(empty) != `{"v":1,"keys":{}}` {
		t.Fatalf("empty = %s, %v", empty, err)
	}
	if got, err := UnmarshalPaths(empty); err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty round trip = %v, %v", got, err)
	}
}

func TestUnmarshalPathsRejects(t *testing.T) {
	for _, in := range []string{`{"v":2,"keys":{}}`, `{"keys":{}}`, `not json`, ``} {
		if _, err := UnmarshalPaths([]byte(in)); err == nil {
			t.Errorf("%q: expected an error", in)
		}
	}
}

func TestRemoved(t *testing.T) {
	managed := Paths{
		"url":             {},
		"auth":            {"token": {}},
		"delivery_policy": {"rate": {}, "groups": {"key": {}, "overrides": {}}},
	}
	tests := []struct {
		name string
		user string
		want []string
	}{
		{name: "nothing removed", user: `{"url":"x","auth":{"token":"t"},"delivery_policy":{"rate":1,"groups":{"key":"k","overrides":{}}}}`},
		{name: "top level", user: `{"auth":{"token":"t"},"delivery_policy":{"rate":1,"groups":{"key":"k","overrides":{}}}}`, want: []string{"url"}},
		{name: "nested", user: `{"url":"x","auth":{"token":"t"},"delivery_policy":{"rate":1,"groups":{"key":"k"}}}`, want: []string{"delivery_policy.groups.overrides"}},
		{name: "whole object", user: `{"url":"x"}`, want: []string{"auth", "delivery_policy"}},
		{name: "explicit null is not removal", user: `{"url":null,"auth":null,"delivery_policy":null}`},
		{name: "empty config", user: `{}`, want: []string{"auth", "delivery_policy", "url"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Removed(managed, mustParse(t, tt.user))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
	if got := Removed(nil, mustParse(t, `{"a":1}`)); got != nil {
		t.Fatalf("no managed keys: %v", got)
	}
}

func TestCanonical(t *testing.T) {
	got, err := Canonical(mustParse(t, `{"b":1152921504606846976,"a":{"z":null,"y":[1.50,"<&>"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":{"y":[1.50,"<&>"],"z":null},"b":1152921504606846976}`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	for _, in := range []string{`null`, `{}`} {
		got, err := Canonical(mustParse(t, in))
		if err != nil || got != "{}" {
			t.Fatalf("%s: got %s, %v", in, got, err)
		}
	}
	if got, err := Canonical(nil); err != nil || got != "{}" {
		t.Fatalf("nil: got %s, %v", got, err)
	}
}

func TestParseRejects(t *testing.T) {
	for _, in := range []string{``, `{`, `{"a":}`, `[1]`, `"x"`, `1`, `{"a":1} {"b":2}`, strings.Repeat("{", 3)} {
		if _, err := Parse(in); err == nil {
			t.Errorf("%q: expected an error", in)
		}
	}
}
