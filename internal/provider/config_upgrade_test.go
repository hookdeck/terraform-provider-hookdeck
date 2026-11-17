package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// upgradeConfig is a source and a destination under the given type names.
// The source sets a key Hookdeck does not return. The destination sets
// authentication, a rate limit written as a string, which Hookdeck stores as
// a number, and optionally its period.
func upgradeConfig(sourceType, destinationType string, period bool) string {
	policy := `{"rate":"10"}`
	if period {
		policy = `{"rate":"10","period":"minute"}`
	}
	return driftObject(sourceType, "mock-upgrade-src", "HTTP", `{"allowed_http_methods":["POST"],"not_a_key":"x"}`) +
		driftObject(destinationType, "mock-upgrade-dst", "HTTP",
			`{"url":"https://mock.hookdeck.com","auth_type":"BEARER_TOKEN","auth":{"token":"secret"},"delivery_policy":`+policy+`}`)
}

func countPuts(m *mockAPI) int {
	count := 0
	for _, prefix := range []string{"/sources/", "/destinations/"} {
		for _, r := range m.requestsTo(prefix) {
			if r.method == http.MethodPut {
				count++
			}
		}
	}
	return count
}

// State written by the released v3 provider holds the configuration as
// written. This version plans nothing for it, although Hookdeck stores a
// value in another form and drops a key, and still clears a key removed
// later. Once this version has written the state, the released provider
// refuses it: it would take every key Hookdeck returned for a configured one
// and clear those the configuration does not have.
func TestAccMock_ConfigUpgradeFromReleasedV3(t *testing.T) {
	for _, tc := range []struct{ name, sourceType, destinationType string }{
		{"gateway names", "hookdeck_gateway_source", "hookdeck_gateway_destination"},
		{"deprecated names", "hookdeck_source", "hookdeck_destination"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := driftMock(t)
			m.update(func() { m.strippedKeys["not_a_key"] = true })
			t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
			source, destination := tc.sourceType+".test", tc.destinationType+".test"
			with := upgradeConfig(tc.sourceType, tc.destinationType, true)
			without := upgradeConfig(tc.sourceType, tc.destinationType, false)
			released := map[string]resource.ExternalProvider{
				"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV3},
			}
			var puts int

			resource.Test(t, resource.TestCase{
				Steps: []resource.TestStep{
					{ExternalProviders: released, Config: requiredProviders(releasedV3) + with},
					{
						ProtoV6ProviderFactories: acc.ProviderFactories,
						Config:                   requiredProviders("") + with,
						ConfigPlanChecks:         emptyPlan(),
					},
					{
						// State now holds what Hookdeck returns.
						ProtoV6ProviderFactories: acc.ProviderFactories,
						Config:                   requiredProviders("") + with,
						ConfigPlanChecks:         emptyPlan(),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttrWith(destination, "config", contains(`"auth":{"token":"secret"}`)),
							resource.TestCheckResourceAttrWith(destination, "config", contains(`"rate":10`)),
							resource.TestCheckResourceAttrWith(destination, "config", contains(`"path_forwarding_disabled":false`)),
							resource.TestCheckResourceAttrWith(source, "config", contains(`"custom_response":null`)),
						),
					},
					{
						ProtoV6ProviderFactories: acc.ProviderFactories,
						Config:                   requiredProviders("") + without,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(source, plancheck.ResourceActionNoop),
							plancheck.ExpectResourceAction(destination, plancheck.ResourceActionUpdate),
						}},
						Check: func(*terraform.State) error {
							sent := lastWrittenConfig(m, "/destinations/")
							policy, _ := sent["delivery_policy"].(map[string]any)
							if v, present := policy["period"]; !present || v != nil {
								return fmt.Errorf("period in the update = %v (present %v), want null", v, present)
							}
							if _, present := sent["http_method"]; present {
								return fmt.Errorf("the update sent a key the configuration never set: %v", sent)
							}
							puts = countPuts(m)
							return nil
						},
					},
					{
						ExternalProviders: released,
						Config:            requiredProviders(releasedV3) + without,
						ExpectError:       regexp.MustCompile(`managed by newer provider version`),
					},
					{
						ProtoV6ProviderFactories: acc.ProviderFactories,
						Config:                   requiredProviders("") + without,
						ConfigPlanChecks:         emptyPlan(),
						Check: func(*terraform.State) error {
							if got := countPuts(m); got != puts {
								return fmt.Errorf("the released provider sent %d updates", got-puts)
							}
							return nil
						},
					},
				},
			})
		})
	}
}

// State the released v3 provider wrote under the deprecated names moves to
// the current names with this version, config and all.
func TestAccMock_ConfigUpgradeFromReleasedV3Moved(t *testing.T) {
	m := driftMock(t)
	m.update(func() { m.strippedKeys["not_a_key"] = true })
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	moved := `
moved {
  from = hookdeck_source.test
  to   = hookdeck_gateway_source.test
}

moved {
  from = hookdeck_destination.test
  to   = hookdeck_gateway_destination.test
}
`

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV3},
				},
				Config: requiredProviders(releasedV3) + upgradeConfig("hookdeck_source", "hookdeck_destination", true),
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + upgradeConfig("hookdeck_gateway_source", "hookdeck_gateway_destination", true) + moved,
				ConfigPlanChecks:         emptyPlan(),
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + upgradeConfig("hookdeck_gateway_source", "hookdeck_gateway_destination", false) + moved,
				ConfigPlanChecks:         updatePlan(destinationAddr),
				Check: func(*terraform.State) error {
					policy, _ := lastWrittenConfig(m, "/destinations/")["delivery_policy"].(map[string]any)
					if v, present := policy["period"]; !present || v != nil {
						return fmt.Errorf("period in the update = %v (present %v), want null", v, present)
					}
					return nil
				},
			},
		},
	})
}
