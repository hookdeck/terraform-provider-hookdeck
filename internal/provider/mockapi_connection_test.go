package provider_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var connectionTypes = []string{"hookdeck_gateway_connection", "hookdeck_connection"}

const delayRules = "  rules = [{ delay_rule = { delay = 1000 } }]\n"

func mockConnectionConfig(resourceType, name, rules string) string {
	return fmt.Sprintf("resource %q \"test\" {\n  name           = %q\n  source_id      = \"src_1\"\n  destination_id = \"dst_1\"\n%s}\n", resourceType, name, rules)
}

// connectionWrites returns the bodies of the create and update requests the
// provider sent for connections, in order.
func (m *mockAPI) connectionWrites() []map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []map[string]any
	for _, r := range m.requests {
		if strings.HasPrefix(r.path, "/connections") && (r.method == http.MethodPost || r.method == http.MethodPut) {
			out = append(out, r.body)
		}
	}
	return out
}

// storedConnections returns the connections the mock API holds. The caller
// holds the lock.
func (m *mockAPI) storedConnections() []map[string]any {
	var out []map[string]any
	for id, item := range m.items {
		if strings.HasPrefix(id, "con_") {
			out = append(out, item)
		}
	}
	return out
}

// checkNoStoredRules checks that the mock API's only connection holds no rules.
// Use it as a step check: destroy removes the connection.
func (m *mockAPI) checkNoStoredRules() resource.TestCheckFunc {
	return func(*terraform.State) error {
		m.mu.Lock()
		defer m.mu.Unlock()
		connections := m.storedConnections()
		if len(connections) != 1 {
			return fmt.Errorf("expected one connection, got %d", len(connections))
		}
		for _, c := range connections {
			if rules, _ := c["rules"].([]any); len(rules) != 0 {
				return fmt.Errorf("expected no rules in the API, got %v", c["rules"])
			}
		}
		return nil
	}
}

func updateAction(address string) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
	}}
}

// requireWrite checks the rules field of the nth write: "absent", "empty"
// (the empty list) or "delay" (one rule).
func requireWrite(t *testing.T, writes []map[string]any, n int, wantRules string) {
	t.Helper()
	if n >= len(writes) {
		t.Fatalf("expected at least %d writes, got %d", n+1, len(writes))
	}
	rules, present := writes[n]["rules"]
	switch wantRules {
	case "absent":
		if present {
			t.Errorf("write %d sent rules %v, expected none: %v", n, rules, writes[n])
		}
	case "empty":
		if list, ok := rules.([]any); !present || !ok || len(list) != 0 {
			t.Errorf("write %d: expected rules [], got %v (present %v): %v", n, rules, present, writes[n])
		}
	case "delay":
		if list, ok := rules.([]any); !present || !ok || len(list) != 1 {
			t.Errorf("write %d: expected one rule, got %v (present %v): %v", n, rules, present, writes[n])
		}
	default:
		t.Fatalf("unknown wantRules %q", wantRules)
	}
}

// A connection that never had rules sends none on create or update, and
// plans nothing.
func TestAccMock_ConnectionWithoutRulesSendsNone(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, address := range connectionTypes {
		t.Run(address, func(t *testing.T) {
			m := mockOrganization(t)
			acc.UseAPIKey(t, mockProjectKeyA)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: mockConnectionConfig(address, "one", ""),
						Check:  resource.TestCheckNoResourceAttr(address+".test", "rules.#"),
					},
					{Config: mockConnectionConfig(address, "one", ""), ConfigPlanChecks: emptyPlan()},
					{
						Config:           mockConnectionConfig(address, "two", ""),
						ConfigPlanChecks: updateAction(address + ".test"),
						Check:            resource.TestCheckNoResourceAttr(address+".test", "rules.#"),
					},
					{Config: mockConnectionConfig(address, "two", ""), ConfigPlanChecks: emptyPlan()},
				},
			})
			writes := m.connectionWrites()
			requireWrite(t, writes, 0, "absent")
			requireWrite(t, writes, 1, "absent")
		})
	}
}

// rules = [] clears the rules and the state keeps the empty list.
func TestAccMock_ConnectionEmptyRulesClearsRules(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, address := range connectionTypes {
		t.Run(address, func(t *testing.T) {
			m := mockOrganization(t)
			acc.UseAPIKey(t, mockProjectKeyA)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: mockConnectionConfig(address, "one", "  rules = []\n"),
						Check:  resource.TestCheckResourceAttr(address+".test", "rules.#", "0"),
					},
					{
						Config: mockConnectionConfig(address, "one", delayRules),
						Check:  resource.TestCheckResourceAttr(address+".test", "rules.#", "1"),
					},
					{
						Config:           mockConnectionConfig(address, "one", "  rules = []\n"),
						ConfigPlanChecks: updateAction(address + ".test"),
						Check:            resource.TestCheckResourceAttr(address+".test", "rules.#", "0"),
					},
					{
						Config:           mockConnectionConfig(address, "one", "  rules = []\n"),
						ConfigPlanChecks: emptyPlan(),
						Check:            m.checkNoStoredRules(),
					},
				},
			})
			writes := m.connectionWrites()
			requireWrite(t, writes, 0, "empty")
			requireWrite(t, writes, 1, "delay")
			requireWrite(t, writes, 2, "empty")
		})
	}
}

// Removing rules from the configuration clears them in Hookdeck and leaves
// null in state, and the next plan is empty.
func TestAccMock_ConnectionRulesRemovedFromConfig(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, address := range connectionTypes {
		t.Run(address, func(t *testing.T) {
			m := mockOrganization(t)
			acc.UseAPIKey(t, mockProjectKeyA)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: mockConnectionConfig(address, "one", delayRules),
						Check:  resource.TestCheckResourceAttr(address+".test", "rules.#", "1"),
					},
					{
						Config:           mockConnectionConfig(address, "one", ""),
						ConfigPlanChecks: updateAction(address + ".test"),
						Check:            resource.TestCheckNoResourceAttr(address+".test", "rules.#"),
					},
					{
						Config:           mockConnectionConfig(address, "one", ""),
						ConfigPlanChecks: emptyPlan(),
						Check:            m.checkNoStoredRules(),
					},
				},
			})
			writes := m.connectionWrites()
			requireWrite(t, writes, 0, "delay")
			requireWrite(t, writes, 1, "empty")
		})
	}
}

// Rules added outside Terraform to a connection configured without any show
// up on refresh and are cleared by the next apply.
func TestAccMock_ConnectionRulesAddedOutsideTerraform(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, address := range connectionTypes {
		t.Run(address, func(t *testing.T) {
			m := mockOrganization(t)
			acc.UseAPIKey(t, mockProjectKeyA)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Steps: []resource.TestStep{
					{Config: mockConnectionConfig(address, "one", "")},
					{
						PreConfig: func() {
							m.update(func() {
								for _, c := range m.storedConnections() {
									c["rules"] = []any{map[string]any{"type": "delay", "delay": float64(500)}}
								}
							})
						},
						Config:           mockConnectionConfig(address, "one", ""),
						ConfigPlanChecks: updateAction(address + ".test"),
						Check:            resource.TestCheckNoResourceAttr(address+".test", "rules.#"),
					},
					{
						Config:           mockConnectionConfig(address, "one", ""),
						ConfigPlanChecks: emptyPlan(),
						Check:            m.checkNoStoredRules(),
					},
				},
			})
			writes := m.connectionWrites()
			requireWrite(t, writes, 0, "absent")
			requireWrite(t, writes, 1, "empty")
		})
	}
}
