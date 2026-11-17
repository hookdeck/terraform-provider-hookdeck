package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// disabled_at is computed only; setting it is a config error.
func TestAccBehaviour_DisabledAtIsReadOnly(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      loadFixture(t, "destination_disabled_at.tf", suffix),
				ExpectError: regexp.MustCompile(`(?s)Read-Only Attribute.*disabled_at`),
			},
		},
	})
}

// Resources deleted outside Terraform drop out of state on refresh and the
// next plan recreates them, instead of the plan failing.
func TestAccBehaviour_DeletedOutsideTerraform(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	paths := map[string]string{
		"hookdeck_gateway_connection.test":     "/connections/",
		"hookdeck_gateway_source.test":         "/sources/",
		"hookdeck_gateway_destination.test":    "/destinations/",
		"hookdeck_gateway_transformation.test": "/transformations/",
	}
	// The connection goes first: it references the others.
	order := []string{
		"hookdeck_gateway_connection.test",
		"hookdeck_gateway_source.test",
		"hookdeck_gateway_destination.test",
		"hookdeck_gateway_transformation.test",
	}
	ids := map[string]string{}

	recreated := []plancheck.PlanCheck{
		plancheck.ExpectResourceAction("hookdeck_gateway_source_auth.test", plancheck.ResourceActionCreate),
	}
	for _, address := range order {
		recreated = append(recreated, plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate))
	}

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "gateway_basic.tf", suffix),
				Check: func(s *terraform.State) error {
					for _, address := range order {
						ids[address] = s.RootModule().Resources[address].Primary.ID
					}
					return nil
				},
			},
			{
				PreConfig: func() {
					for _, address := range order {
						status, body := acc.Request(t, acc.OrgKey(t), acc.ProjectA(t), "DELETE", paths[address]+ids[address], nil)
						if status > 299 {
							t.Fatalf("delete %s %s: %d %v", address, ids[address], status, body)
						}
					}
				},
				Config:           loadFixture(t, "gateway_basic.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: recreated},
			},
		},
	})
}

// Rules removed from a connection outside Terraform show as drift on refresh
// and the next apply puts them back.
func TestAccBehaviour_ConnectionRulesRemovedOutsideTerraform(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_connection.test"
	var id string

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "gateway_basic.tf", suffix),
				Check: func(s *terraform.State) error {
					id = s.RootModule().Resources[address].Primary.ID
					return nil
				},
			},
			{
				PreConfig: func() {
					status, body := acc.Request(t, acc.OrgKey(t), acc.ProjectA(t), "PUT", "/connections/"+id, map[string]any{"rules": []any{}})
					if status > 299 {
						t.Fatalf("clear rules of %s: %d %v", id, status, body)
					}
				},
				Config: loadFixture(t, "gateway_basic.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "rules.#", "1"),
					func(*terraform.State) error {
						status, body, err := acc.RequestCurrent("GET", "/connections/"+id, nil, nil)
						if err != nil {
							return err
						}
						if rules, _ := body["rules"].([]any); status > 299 || len(rules) != 1 {
							return fmt.Errorf("connection %s after apply: status %d, rules %v", id, status, body["rules"])
						}
						return nil
					},
				),
			},
		},
	})
}

// Removing rules from a connection's configuration clears them in Hookdeck.
func TestAccBehaviour_ConnectionRulesRemovedFromConfig(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_connection.test"
	withRules := loadFixture(t, "gateway_basic.tf", suffix)
	withoutRules := regexp.MustCompile(`(?s)\n  rules = \[\{.*?\}\]\n`).ReplaceAllString(withRules, "\n")
	if withoutRules == withRules {
		t.Fatal("gateway_basic.tf: rules block not matched, nothing removed")
	}
	var id string

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: withRules,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "rules.#", "1"),
					func(s *terraform.State) error {
						id = s.RootModule().Resources[address].Primary.ID
						return nil
					},
				),
			},
			{
				Config: withoutRules,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(address, "rules.#"),
					func(*terraform.State) error {
						status, body, err := acc.RequestCurrent("GET", "/connections/"+id, nil, nil)
						if err != nil {
							return err
						}
						if rules, _ := body["rules"].([]any); status > 299 || len(rules) != 0 {
							return fmt.Errorf("connection %s after apply: status %d, rules %v", id, status, body["rules"])
						}
						return nil
					},
				),
			},
			{Config: withoutRules, ConfigPlanChecks: emptyPlan()},
		},
	})
}
