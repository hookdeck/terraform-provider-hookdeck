package destination

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// deliveryPolicyValidator catches retired fields during planning, including when
// an unchanged legacy configuration would otherwise produce an empty plan.
type deliveryPolicyValidator struct{}

func (deliveryPolicyValidator) Description(context.Context) string {
	return "Destination config must use the 2026-09-01 delivery_policy format."
}

func (v deliveryPolicyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (deliveryPolicyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &config); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid destination config", "Destination config must be a JSON object.")
		return
	}
	if err := validateDeliveryPolicy(config); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Destination config migration required", err.Error())
	}
}

func validateDeliveryPolicy(config map[string]interface{}) error {
	for _, field := range []struct{ old, replacement string }{
		{"rate_limit", "delivery_policy.rate"},
		{"rate_limit_period", "delivery_policy.period"},
		{"delivery_groups", "delivery_policy.groups"},
	} {
		if _, exists := config[field.old]; exists {
			return fmt.Errorf("config.%s is no longer supported by API 2026-09-01; update your Terraform configuration to config.%s (no automatic translation is performed). See the destination delivery policy migration guide", field.old, field.replacement)
		}
	}
	policy, _ := config["delivery_policy"].(map[string]interface{})
	groups, _ := policy["groups"].(map[string]interface{})
	if err := validateGroupRateFields(groups); err != nil {
		return err
	}
	overrides, _ := groups["overrides"].(map[string]interface{})
	for _, value := range overrides {
		if override, ok := value.(map[string]interface{}); ok {
			if err := validateGroupRateFields(override); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateGroupRateFields(group map[string]interface{}) error {
	for _, field := range []struct{ old, replacement string }{{"rate_limit", "rate"}, {"rate_limit_period", "rate_period"}} {
		if _, exists := group[field.old]; exists {
			return fmt.Errorf("delivery_policy.groups and its overrides must use %s instead of %s in API 2026-09-01; update your Terraform configuration", field.replacement, field.old)
		}
	}
	return nil
}
