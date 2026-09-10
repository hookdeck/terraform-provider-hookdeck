package destination_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDestinationResource_RejectsRemovedConfigFields(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-legacy")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`resource "hookdeck_destination" "test" {
 name = %q
 type = "HTTP"
 config = jsonencode({url = "https://mock.hookdeck.com", rate_limit = 10, rate_limit_period = "minute"})
}`, name),
			ExpectError: regexp.MustCompile(`(?s)Destination config uses removed fields.*rate_limit, rate_limit_period`),
		}},
	})
}

func TestAccDestinationResource_DeliveryPolicy(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-policy")
	config := func(fields ...string) string {
		all := append([]string{`url = "https://mock.hookdeck.com"`}, fields...)
		return fmt.Sprintf(`resource "hookdeck_destination" "test" {
 name = %q
 type = "HTTP"
 config = jsonencode({%s})
}`, name, strings.Join(all, ", "))
	}
	checkGroups := func(expected string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs := s.RootModule().Resources["hookdeck_destination.test"]
			if rs == nil {
				return fmt.Errorf("missing destination")
			}
			actual, err := fetchDestinationConfig(rs.Primary.ID)
			if err != nil {
				return err
			}
			policy, _ := actual["delivery_policy"].(map[string]interface{})
			var want interface{}
			if err := json.Unmarshal([]byte(expected), &want); err != nil {
				return err
			}
			if !reflect.DeepEqual(policy["groups"], want) {
				return fmt.Errorf("groups: got %v, want %v", policy["groups"], want)
			}
			return nil
		}
	}
	withGroups := `delivery_policy = {rate = 10, period = "second", groups = {key = "body.customer_id", rate = 5, rate_period = "second", overrides = {priority = {rate = 8, rate_period = "minute"}}}}`
	withoutOverride := `delivery_policy = {rate = 10, period = "second", groups = {key = "body.customer_id", rate = 5, rate_period = "second"}}`
	rateOnly := `delivery_policy = {rate = 10, period = "second"}`
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(rateOnly), Check: checkAPIConfigValue("hookdeck_destination.test", "delivery_policy.rate", float64(10))},
			{Config: config(withGroups), Check: checkGroups(`{"key":"body.customer_id","rate":5,"rate_period":"second","overrides":{"priority":{"rate":8,"rate_period":"minute"}}}`)},
			{Config: config(withoutOverride), Check: func(s *terraform.State) error {
				actual, err := fetchDestinationConfig(s.RootModule().Resources["hookdeck_destination.test"].Primary.ID)
				if err != nil {
					return err
				}
				policy, _ := actual["delivery_policy"].(map[string]interface{})
				groups, ok := policy["groups"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("missing groups")
				}
				if overrides, ok := groups["overrides"].(map[string]interface{}); ok && len(overrides) > 0 {
					return fmt.Errorf("removed overrides retained: %v", overrides)
				}
				return nil
			}},
			{Config: config(rateOnly), Check: checkGroups(`null`)},
			{Config: config(), Check: checkAPIConfigValue("hookdeck_destination.test", "delivery_policy.rate", nil)},
			{Config: config(), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
