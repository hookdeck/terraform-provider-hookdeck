package source_test

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

// source returns a hookdeck_gateway_source named test; an empty config leaves the attribute out.
func source(rName, sourceType, config string) string {
	attribute := ""
	if config != "" {
		attribute = fmt.Sprintf("  config = jsonencode(%s)\n", config)
	}
	return fmt.Sprintf("resource \"hookdeck_gateway_source\" \"test\" {\n  name = \"tf-acc-source-%s\"\n  type = %q\n%s}\n", rName, sourceType, attribute)
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

// Keys Hookdeck adds to the config are not drift, and a key set in the
// configuration that is changed outside Terraform is a change that the apply
// restores.
func TestAccSourceResource_ConfigDrift(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_source.test"
	config := source(rName, "HTTP", `{ allowed_http_methods = ["GET", "POST"] }`)
	var id string

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: config, Check: capture(address, &id)},
			{
				RefreshState: true,
				Check:        checkConfigContains(address, `"allowed_http_methods":["GET","POST"]`, `"custom_response"`),
			},
			{Config: config, ConfigPlanChecks: emptyPlan()},
			{
				PreConfig: func() {
					status, body, err := acc.RequestCurrent("PUT", "/sources/"+id, nil, map[string]any{
						"type":   "HTTP",
						"config": map[string]any{"allowed_http_methods": []string{"PUT"}},
					})
					if err != nil || status > 299 {
						t.Fatalf("PUT source %s: %d %v %v", id, status, body, err)
					}
				},
				Config:           config,
				ConfigPlanChecks: updatePlan(address),
				Check:            checkAPIConfigStringSlice(address, "allowed_http_methods", []string{"GET", "POST"}),
			},
		},
	})
}

// Changing the type is an update; the configuration is compared against the
// new type's config afterwards.
func TestAccSourceResource_ConfigTypeChange(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_source.test"
	webhook := source(rName, "WEBHOOK", `{ allowed_http_methods = ["POST"] }`)
	httpSource := source(rName, "HTTP", `{ allowed_http_methods = ["POST"] }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: webhook},
			{Config: httpSource, ConfigPlanChecks: updatePlan(address)},
			{Config: httpSource, ConfigPlanChecks: emptyPlan()},
		},
	})
}

// A source created outside Terraform and imported plans nothing when the
// configuration holds a subset of what it has.
func TestAccSourceResource_ConfigImport(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	address := "hookdeck_gateway_source.test"
	var id string
	t.Cleanup(func() {
		if id != "" {
			_, _, _ = acc.RequestCurrent("DELETE", "/sources/"+id, nil, nil)
		}
	})
	matching := source(rName, "HTTP", `{ allowed_http_methods = ["POST"] }`)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					status, body, err := acc.RequestCurrent("POST", "/sources", nil, map[string]any{
						"name":   "tf-acc-source-" + rName,
						"type":   "HTTP",
						"config": map[string]any{"allowed_http_methods": []string{"POST"}},
					})
					if err != nil || status > 299 {
						t.Fatalf("POST source: %d %v %v", status, body, err)
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
		},
	})
}
