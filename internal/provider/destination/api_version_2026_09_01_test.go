package destination

import (
	"reflect"
	"testing"
)

func TestRemovedConfigFieldsIn20260901(t *testing.T) {
	tests := []struct {
		name, config string
		want         []string
	}{
		{name: "new shape", config: `{"url":"https://example.com","delivery_policy":{"rate":1}}`},
		{name: "all removed", config: `{"rate_limit":1,"rate_limit_period":"second","delivery_groups":{"key":"body.id"}}`, want: []string{"rate_limit", "rate_limit_period", "delivery_groups"}},
		{name: "removed set to null still flagged", config: `{"rate_limit":null}`, want: []string{"rate_limit"}},
		{name: "nested names ignored", config: `{"delivery_policy":{"groups":{"rate_limit":1}}}`},
		{name: "malformed", config: `{`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removedConfigFieldsIn20260901(tt.config); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
