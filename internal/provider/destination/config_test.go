package destination_test

import (
	"fmt"
	"strings"
	"terraform-provider-hookdeck/internal/acc"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func emptyPlan() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}
}

func updatePlan(address string) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
	}}
}

// destination returns a destination named test with the config given as HCL.
func destination(resourceType, rName, config string) string {
	return fmt.Sprintf(`resource %[1]q "test" {
  name   = "tf-acc-destination-%[2]s"
  type   = "HTTP"
  config = jsonencode(%[3]s)
}
`, resourceType, rName, config)
}

func capture(address string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[address]
		if !ok {
			return fmt.Errorf("resource not found: %s", address)
		}
		*into = rs.Primary.ID
		return nil
	}
}

// putConfig changes the destination's config the way an edit outside
// Terraform does.
func putConfig(t *testing.T, id string, config map[string]any) {
	t.Helper()
	status, body, err := acc.RequestCurrent("PUT", "/destinations/"+id, nil, map[string]any{"config": config})
	if err != nil || status > 299 {
		t.Fatalf("PUT destination %s: %d %v %v", id, status, body, err)
	}
}

func checkConfigContains(address string, wants ...string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(address, "config", func(value string) error {
		for _, want := range wants {
			if !strings.Contains(value, want) {
				return fmt.Errorf("config %s does not contain %s", value, want)
			}
		}
		return nil
	})
}

// Keys Hookdeck adds to the config are not drift: the plan is empty after
// apply and after a refresh stores the full config.
func TestAccDestinationResource_ConfigDefaultsAreNotDrift(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_destination.test"
	config := destination("hookdeck_gateway_destination", rName, `{ url = "https://mock.hookdeck.com" }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: config},
			{
				RefreshState: true,
				Check:        checkConfigContains(address, `"url":"https://mock.hookdeck.com"`, `"path_forwarding_disabled"`, `"auth_type"`),
			},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A key set in the configuration that is changed outside Terraform is a
// change and the apply restores it. A key that is not set is not compared.
func TestAccDestinationResource_ConfigDrift(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_destination.test"
	config := destination("hookdeck_destination", rName, `{
    url = "https://mock.hookdeck.com"
    delivery_policy = {
      rate   = 10
      period = "second"
    }
  }`)
	var id string

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: config, Check: capture(address, &id)},
			{
				PreConfig: func() {
					putConfig(t, id, map[string]any{"delivery_policy": map[string]any{"rate": 20}})
				},
				Config:           config,
				ConfigPlanChecks: updatePlan(address),
				Check:            checkAPIConfigValue(address, "delivery_policy.rate", float64(10)),
			},
			{
				PreConfig:        func() { putConfig(t, id, map[string]any{"http_method": "PUT"}) },
				Config:           config,
				ConfigPlanChecks: emptyPlan(),
			},
		},
	})
}

// A secret set in the configuration is compared in clear.
func TestAccDestinationResource_ConfigWithAuth(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_destination.test"
	config := destination("hookdeck_gateway_destination", rName, `{
    url       = "https://mock.hookdeck.com"
    auth_type = "BEARER_TOKEN"
    auth = {
      token = "secret-`+rName+`"
    }
  }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: config},
			{RefreshState: true, Check: checkConfigContains(address, `"auth_type":"BEARER_TOKEN"`)},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A number written as a string comes back as a number.
func TestAccDestinationResource_ConfigNumberAsString(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := destination("hookdeck_gateway_destination", rName, `{
    url = "https://mock.hookdeck.com"
    delivery_policy = {
      rate   = "10"
      period = "second"
    }
  }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: config},
			{Config: config, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A destination created outside Terraform and imported plans nothing when the
// configuration holds a subset of what it has.
func TestAccDestinationResource_ConfigImport(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_destination.test"
	var id string
	t.Cleanup(func() {
		if id != "" {
			_, _, _ = acc.RequestCurrent("DELETE", "/destinations/"+id, nil, nil)
		}
	})
	matching := destination("hookdeck_gateway_destination", rName, `{ url = "https://mock.hookdeck.com" }`)
	differing := destination("hookdeck_gateway_destination", rName, `{ url = "https://mock.hookdeck.com/other" }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					status, body, err := acc.RequestCurrent("POST", "/destinations", nil, map[string]any{
						"name":   "tf-acc-destination-" + rName,
						"type":   "HTTP",
						"config": map[string]any{"url": "https://mock.hookdeck.com"},
					})
					if err != nil || status > 299 {
						t.Fatalf("POST destination: %d %v %v", status, body, err)
					}
					id, _ = body["id"].(string)
				},
				Config:             matching,
				ResourceName:       address,
				ImportState:        true,
				ImportStateIdFunc:  func(*terraform.State) (string, error) { return id, nil },
				ImportStatePersist: true,
			},
			{Config: matching, ConfigPlanChecks: emptyPlan()},
			{Config: differing, ConfigPlanChecks: updatePlan(address)},
		},
	})
}
