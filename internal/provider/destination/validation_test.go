package destination

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDeliveryPolicyValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  types.String
		invalid bool
	}{
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
		{"no policy", types.StringValue(`{"url":"https://example.com"}`), false},
		{"new policy", types.StringValue(`{"delivery_policy":{"rate":10,"period":"minute"}}`), false},
		{"old rate", types.StringValue(`{"rate_limit":10}`), true},
		{"old null rate", types.StringValue(`{"rate_limit":null}`), true},
		{"old period", types.StringValue(`{"rate_limit_period":"second"}`), true},
		{"old groups", types.StringValue(`{"delivery_groups":{}}`), true},
		{"mixed", types.StringValue(`{"rate_limit":10,"delivery_policy":{"rate":10}}`), true},
		{"old group rate", types.StringValue(`{"delivery_policy":{"groups":{"rate_limit":10}}}`), true},
		{"old override period", types.StringValue(`{"delivery_policy":{"groups":{"overrides":{"a":{"rate_limit_period":"second"}}}}}`), true},
		{"new groups", types.StringValue(`{"delivery_policy":{"groups":{"key":"body.id","rate":5,"rate_period":"second","overrides":{"a":{"rate":2,"rate_period":"minute"}}}}}`), false},
		{"malformed", types.StringValue(`{`), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp validator.StringResponse
			deliveryPolicyValidator{}.ValidateString(t.Context(), validator.StringRequest{Path: path.Root("config"), ConfigValue: tt.config}, &resp)
			if resp.Diagnostics.HasError() != tt.invalid {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
		})
	}
}
