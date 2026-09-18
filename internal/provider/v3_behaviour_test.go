package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// disabled_at was accepted in v2 and never sent. In v3 it is computed only.
func TestAccV3_DisabledAtIsReadOnly(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      loadFixture(t, "destination_disabled_at.tf", suffix),
				ExpectError: regexp.MustCompile(`disabled_at`),
			},
		},
	})
}

// A resource deleted outside Terraform drops out of state on refresh and
// the next plan recreates it, instead of the plan failing.
func TestAccV3_DeletedOutsideTerraform(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var sourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "source_without_project_id.tf", suffix),
				Check:  captureID("hookdeck_gateway_source.test", &sourceID),
			},
			{
				PreConfig: func() {
					status, body := apiRequest(t, os.Getenv(envAPIKey), "DELETE", "/2026-09-01/sources/"+sourceID)
					if status > 299 {
						t.Fatalf("delete source %s: %d %v", sourceID, status, body)
					}
				},
				Config: loadFixture(t, "source_without_project_id.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("hookdeck_gateway_source.test", plancheck.ResourceActionCreate),
					},
				},
				Check: func(s *terraform.State) error {
					rs := s.RootModule().Resources["hookdeck_gateway_source.test"]
					if rs == nil {
						return fmt.Errorf("source not recreated")
					}
					if rs.Primary.ID == sourceID {
						return fmt.Errorf("source id %s unchanged after out-of-band delete", sourceID)
					}
					return nil
				},
			},
		},
	})
}

// --- placeholders for RFC items with their own discussion ---

func TestAccV3_ConnectionDisabledAndPaused(t *testing.T) {
	t.Skip("RFC #229 'Connection disabled and paused' (#65, #128): not part of this PR")
}

func TestAccV3_ConfigRefreshedFromAPI(t *testing.T) {
	t.Skip("RFC #229 'config refresh' (#144, #228): not part of this PR")
}
