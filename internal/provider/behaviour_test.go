package provider_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// disabled_at is computed only; setting it is a config error.
func TestAccBehaviour_DisabledAtIsReadOnly(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
						status, body := apiRequest(t, os.Getenv(envAPIKey), "", "DELETE", paths[address]+ids[address], nil)
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
