package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const destinationAddr = "hookdeck_gateway_destination.test"

// driftSource returns a source named test with the given config JSON. An empty
// config leaves the attribute out.
func driftSource(resourceType, config string) string {
	return driftObject(resourceType, "mock-drift-src", "HTTP", config)
}

func driftDestination(resourceType, config string) string {
	return driftObject(resourceType, "mock-drift-dst", "HTTP", config)
}

func driftObject(resourceType, name, objectType, config string) string {
	line := ""
	if config != "" {
		line = fmt.Sprintf("  config = %q\n", config)
	}
	return fmt.Sprintf("resource %q \"test\" {\n  name = %q\n  type = %q\n%s}\n", resourceType, name, objectType, line)
}

func updatePlan(addr string) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
	}}
}

// lastWrittenConfig returns the config of the last PUT to the objects under prefix.
func lastWrittenConfig(m *mockAPI, prefix string) map[string]any {
	var config map[string]any
	for _, r := range m.requestsTo(prefix) {
		if r.method == http.MethodPut {
			config, _ = r.body["config"].(map[string]any)
		}
	}
	return config
}

// storedConfig returns the config the mock holds for an object, as JSON.
func storedConfig(m *mockAPI, id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, store := range []map[string]map[string]any{m.sources, m.destinations} {
		if object, ok := store[id]; ok {
			b, _ := json.Marshal(object["config"])
			return string(b)
		}
	}
	return ""
}

func checkStoredConfig(m *mockAPI, check func(config map[string]any) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[sourceAddr]
		if !ok {
			return fmt.Errorf("resource not found: %s", sourceAddr)
		}
		var config map[string]any
		if err := json.Unmarshal([]byte(storedConfig(m, rs.Primary.ID)), &config); err != nil {
			return err
		}
		return check(config)
	}
}

func driftMock(t *testing.T) *mockAPI {
	t.Helper()
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockProjectKeyA)
	return m
}

// Hookdeck adds defaults to the config: the plan stays empty after apply and
// after the refresh that stores them in state.
func TestAccMock_ConfigDefaultsAreNotDrift(t *testing.T) {
	driftMock(t)
	config := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith(sourceAddr, "config", contains(`"custom_response":null`)),
					resource.TestCheckResourceAttrWith(sourceAddr, "config", contains(`"allowed_http_methods":["POST"]`)),
				),
			},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

func contains(want string) resource.CheckResourceAttrWithFunc {
	return func(value string) error {
		if !strings.Contains(value, want) {
			return fmt.Errorf("%s does not contain %s", value, want)
		}
		return nil
	}
}

// A key set in the configuration that is changed outside Terraform is a
// change, and the apply restores it.
func TestAccMock_ConfigDriftOnManagedKey(t *testing.T) {
	m := driftMock(t)
	config := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)
	var id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(sourceAddr, &id)},
			{
				PreConfig:        func() { m.setSourceConfig(id, "allowed_http_methods", []any{"GET"}) },
				Config:           config,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: checkStoredConfig(m, func(c map[string]any) error {
					if !reflect.DeepEqual(c["allowed_http_methods"], []any{"POST"}) {
						return fmt.Errorf("allowed_http_methods = %v", c["allowed_http_methods"])
					}
					return nil
				}),
			},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A key that is not in the configuration is not compared.
func TestAccMock_ConfigDriftOnUnmanagedKeyIsIgnored(t *testing.T) {
	m := driftMock(t)
	config := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)
	var id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(sourceAddr, &id)},
			{
				PreConfig: func() {
					m.setSourceConfig(id, "custom_response", map[string]any{"content_type": "json", "body": "x"})
				},
				Config:           config,
				ConfigPlanChecks: emptyPlan(),
			},
		},
	})
}

// Removing a key from the configuration clears it in Hookdeck, as before
// the config was read back.
func TestAccMock_ConfigRemovedKeyIsCleared(t *testing.T) {
	m := driftMock(t)
	with := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`)
	without := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: with},
			{
				Config:           without,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(*terraform.State) error {
						sent := lastWrittenConfig(m, "/sources/")
						if v, present := sent["custom_response"]; !present || v != nil {
							return fmt.Errorf("custom_response in the update = %v (present %v), want null", v, present)
						}
						return nil
					},
					checkStoredConfig(m, func(c map[string]any) error {
						if c["custom_response"] != nil {
							return fmt.Errorf("custom_response = %v", c["custom_response"])
						}
						return nil
					}),
				),
			},
			{Config: without, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// Removing the whole config clears every key that was set.
func TestAccMock_ConfigRemovedEntirelyIsCleared(t *testing.T) {
	m := driftMock(t)
	with := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`)
	without := driftSource("hookdeck_gateway_source", "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: with},
			{
				Config:           without,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: func(*terraform.State) error {
					sent := lastWrittenConfig(m, "/sources/")
					if v, present := sent["custom_response"]; !present || v != nil {
						return fmt.Errorf("custom_response in the update = %v (present %v), want null", v, present)
					}
					if !reflect.DeepEqual(sent["allowed_http_methods"], []any{"PUT", "POST", "PATCH", "DELETE"}) {
						return fmt.Errorf("allowed_http_methods in the update = %v, want the default", sent["allowed_http_methods"])
					}
					return nil
				},
			},
			{Config: without, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A key Hookdeck drops (it is not valid for the type) does not plan an
// update on every run.
func TestAccMock_ConfigKeyDroppedByHookdeck(t *testing.T) {
	m := driftMock(t)
	m.update(func() { m.strippedKeys["not_a_key"] = true })
	config := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"not_a_key":"x"}`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{Config: config, ConfigPlanChecks: emptyPlan()},
			{RefreshState: true, Check: resource.TestCheckResourceAttrWith(sourceAddr, "config", func(v string) error {
				if strings.Contains(v, "not_a_key") {
					return fmt.Errorf("state holds the dropped key: %s", v)
				}
				return nil
			})},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

func hasNullCustomResponse(m *mockAPI) resource.TestCheckFunc {
	return func(*terraform.State) error {
		sent := lastWrittenConfig(m, "/sources/")
		if v, present := sent["custom_response"]; !present || v != nil {
			return fmt.Errorf("custom_response in the update = %v (present %v), want null", v, present)
		}
		return nil
	}
}

// A source created outside Terraform and imported plans nothing when the
// configuration matches. Removing a key does nothing until an apply has
// recorded which keys the configuration sets; a key set to null clears it
// straight away.
func TestAccMock_ConfigImportRemovedKey(t *testing.T) {
	m := driftMock(t)
	id := m.createSource("tm_a", "mock-drift-src", "HTTP", map[string]any{
		"allowed_http_methods": []any{"POST"},
		"custom_response":      map[string]any{"content_type": "json", "body": "x"},
	})
	with := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`)
	without := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)
	cleared := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":null}`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             with,
				ResourceName:       sourceAddr,
				ImportState:        true,
				ImportStateId:      id,
				ImportStatePersist: true,
			},
			{Config: with, ConfigPlanChecks: emptyPlan()},
			// The documented edge: nothing was recorded as set, so a removed key
			// is not planned.
			{Config: without, ConfigPlanChecks: emptyPlan()},
			{
				Config:           cleared,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: resource.ComposeAggregateTestCheckFunc(
					hasNullCustomResponse(m),
					checkStoredConfig(m, func(c map[string]any) error {
						if c["custom_response"] != nil {
							return fmt.Errorf("custom_response = %v", c["custom_response"])
						}
						return nil
					}),
				),
			},
			{Config: cleared, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// An imported source plans an update when a key the configuration sets differs.
func TestAccMock_ConfigImportDiffering(t *testing.T) {
	m := driftMock(t)
	id := m.createSource("tm_a", "mock-drift-src", "HTTP", map[string]any{"allowed_http_methods": []any{"POST"}})
	matching := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)
	differing := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["PUT"]}`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             matching,
				ResourceName:       sourceAddr,
				ImportState:        true,
				ImportStateId:      id,
				ImportStatePersist: true,
			},
			{Config: matching, ConfigPlanChecks: emptyPlan()},
			{
				Config:           differing,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: checkStoredConfig(m, func(c map[string]any) error {
					if !reflect.DeepEqual(c["allowed_http_methods"], []any{"PUT"}) {
						return fmt.Errorf("allowed_http_methods = %v", c["allowed_http_methods"])
					}
					return nil
				}),
			},
			{Config: differing, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A key added with the value Hookdeck already holds is recorded by an apply,
// so removing it later clears it.
func TestAccMock_ConfigAddedKeyWithMatchingValueIsRecorded(t *testing.T) {
	m := driftMock(t)
	without := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`)
	with := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`)
	var id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: without, Check: captureID(sourceAddr, &id)},
			{
				// Someone sets the key in Hookdeck.
				PreConfig: func() {
					m.setSourceConfig(id, "custom_response", map[string]any{"content_type": "json", "body": "x"})
				},
				Config:           with,
				ConfigPlanChecks: updatePlan(sourceAddr),
			},
			{Config: with, ConfigPlanChecks: emptyPlan()},
			{
				Config:           without,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: resource.ComposeAggregateTestCheckFunc(
					hasNullCustomResponse(m),
					checkStoredConfig(m, func(c map[string]any) error {
						if c["custom_response"] != nil {
							return fmt.Errorf("custom_response = %v", c["custom_response"])
						}
						return nil
					}),
				),
			},
			{Config: without, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A configuration that matches Hookdeck is no change by itself, and does not
// hide another change made in the same plan.
func TestAccMock_ConfigMatchingDoesNotHideOtherChanges(t *testing.T) {
	m := driftMock(t)
	config := `{"allowed_http_methods":["POST"]}`
	plain := driftSource("hookdeck_gateway_source", config)
	described := fmt.Sprintf("resource \"hookdeck_gateway_source\" \"test\" {\n  name = \"mock-drift-src\"\n  type = \"HTTP\"\n  description = \"new\"\n  config = %q\n}\n", config)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: plain},
			{RefreshState: true},
			{Config: plain, ConfigPlanChecks: emptyPlan()},
			{
				Config:           described,
				ConfigPlanChecks: updatePlan(sourceAddr),
				Check: func(s *terraform.State) error {
					id := s.RootModule().Resources[sourceAddr].Primary.ID
					m.mu.Lock()
					defer m.mu.Unlock()
					if got := m.sources[id]["description"]; got != "new" {
						return fmt.Errorf("description = %v", got)
					}
					return nil
				},
			},
			{Config: described, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// The data source returns the config Hookdeck holds, defaults included.
func TestAccMock_ConfigDataSources(t *testing.T) {
	m := driftMock(t)
	sourceID := m.createSource("tm_a", "mock-drift-src", "HTTP", map[string]any{"allowed_http_methods": []any{"POST"}})
	destinationID := m.createDestination("tm_a", "mock-drift-dst", "HTTP", map[string]any{"url": "https://mock.hookdeck.com"})
	config := fmt.Sprintf(`data "hookdeck_gateway_source" "test" {
  id = %q
}

data "hookdeck_gateway_destination" "test" {
  id = %q
}
`, sourceID, destinationID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_source.test", "config", contains(`"custom_response":null`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_source.test", "config", contains(`"allowed_http_methods":["POST"]`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_destination.test", "config", contains(`"url":"https://mock.hookdeck.com"`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_destination.test", "config", contains(`"auth_type":"HOOKDECK_SIGNATURE"`)),
				),
			},
		},
	})
}

// Destinations compare the keys set, including members of the delivery
// policy, and clear members that are removed.
func TestAccMock_ConfigDestination(t *testing.T) {
	m := driftMock(t)
	both := driftDestination("hookdeck_gateway_destination", `{"url":"https://mock.hookdeck.com","delivery_policy":{"rate":5,"period":"second"}}`)
	rateOnly := driftDestination("hookdeck_gateway_destination", `{"url":"https://mock.hookdeck.com","delivery_policy":{"rate":5}}`)
	var id string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: both, Check: captureID(destinationAddr, &id)},
			{Config: both, ConfigPlanChecks: emptyPlan()},
			{
				PreConfig: func() {
					m.setDestinationConfig(id, "delivery_policy", map[string]any{"rate": 9, "period": "second", "groups": nil})
				},
				Config:           both,
				ConfigPlanChecks: updatePlan(destinationAddr),
			},
			{
				Config:           rateOnly,
				ConfigPlanChecks: updatePlan(destinationAddr),
				Check: func(*terraform.State) error {
					sent := lastWrittenConfig(m, "/destinations/")
					policy, _ := sent["delivery_policy"].(map[string]any)
					if v, present := policy["period"]; !present || v != nil {
						return fmt.Errorf("period in the update = %v (present %v), want null", v, present)
					}
					return nil
				},
			},
			{Config: rateOnly, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// State written by v2.4.0, which holds exactly the configuration: this
// version plans nothing, and still clears a key that is later removed.
func TestAccMock_ConfigUpgradeFromReleasedV2(t *testing.T) {
	m := driftMock(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	with := driftSource("hookdeck_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`) +
		driftDestination("hookdeck_destination", `{"url":"https://mock.hookdeck.com"}`)
	without := driftSource("hookdeck_source", `{"allowed_http_methods":["POST"]}`) +
		driftDestination("hookdeck_destination", `{"url":"https://mock.hookdeck.com"}`)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV2},
				},
				Config: requiredProviders(releasedV2) + with,
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + with,
				ConfigPlanChecks:         emptyPlan(),
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + without,
				ConfigPlanChecks:         updatePlan("hookdeck_source.test"),
				Check: func(*terraform.State) error {
					sent := lastWrittenConfig(m, "/sources/")
					if v, present := sent["custom_response"]; !present || v != nil {
						return fmt.Errorf("custom_response in the update = %v (present %v), want null", v, present)
					}
					return nil
				},
			},
		},
	})
}

// The record of which keys were set survives a moved block.
func TestAccMock_ConfigKeysSurviveMove(t *testing.T) {
	m := driftMock(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	with := driftSource("hookdeck_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`)
	moved := `
moved {
  from = hookdeck_source.test
  to   = hookdeck_gateway_source.test
}
`
	withMoved := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"],"custom_response":{"content_type":"json","body":"x"}}`) + moved
	withoutMoved := driftSource("hookdeck_gateway_source", `{"allowed_http_methods":["POST"]}`) + moved

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		Steps: []resource.TestStep{
			{ProtoV6ProviderFactories: acc.ProviderFactories, Config: with},
			{ProtoV6ProviderFactories: acc.ProviderFactories, Config: withMoved, ConfigPlanChecks: emptyPlan()},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   withoutMoved,
				ConfigPlanChecks:         updatePlan(sourceAddr),
				Check: func(*terraform.State) error {
					sent := lastWrittenConfig(m, "/sources/")
					if v, present := sent["custom_response"]; !present || v != nil {
						return fmt.Errorf("custom_response in the update = %v (present %v), want null", v, present)
					}
					return nil
				},
			},
		},
	})
}
