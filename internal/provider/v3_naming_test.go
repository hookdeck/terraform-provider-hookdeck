package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Every Event Gateway resource is reachable under its gateway_ name and
// round-trips through import.
func TestAccV3_GatewayResourceNames(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "gateway_basic.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "name", "v3-src-"+suffix),
					resource.TestCheckResourceAttrSet("hookdeck_gateway_source.test", "url"),
					resource.TestCheckResourceAttr("hookdeck_gateway_source_auth.test", "auth_type", "API_KEY"),
					resource.TestCheckResourceAttr("hookdeck_gateway_destination.test", "name", "v3-dst-"+suffix),
					resource.TestCheckResourceAttr("hookdeck_gateway_transformation.test", "name", "v3-trs-"+suffix),
					resource.TestCheckResourceAttr("hookdeck_gateway_connection.test", "name", "v3-con-"+suffix),
					resource.TestCheckResourceAttrPair("hookdeck_gateway_connection.test", "source_id", "hookdeck_gateway_source.test", "id"),
				),
			},
			{
				ResourceName:      "hookdeck_gateway_source.test",
				ImportState:       true,
				ImportStateVerify: true,
				// source_auth updates the source after creation, so
				// updated_at in state predates the import.
				ImportStateVerifyIgnore: []string{"config", "updated_at"},
			},
			{
				ResourceName:            "hookdeck_gateway_destination.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				ResourceName:      "hookdeck_gateway_transformation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "hookdeck_gateway_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// A v2 configuration renamed with moved blocks plans as no-op and keeps
// every resource id.
func TestAccV3_MovedFromV2Names(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var sourceID, destinationID, transformationID, connectionID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "v2_names.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID("hookdeck_source.test", &sourceID),
					captureID("hookdeck_destination.test", &destinationID),
					captureID("hookdeck_transformation.test", &transformationID),
					captureID("hookdeck_connection.test", &connectionID),
				),
			},
			{
				Config: loadFixture(t, "v3_names_moved.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectResourceAction("hookdeck_gateway_source.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("hookdeck_gateway_destination.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("hookdeck_gateway_transformation.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("hookdeck_gateway_connection.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals("hookdeck_gateway_source.test", &sourceID),
					checkIDEquals("hookdeck_gateway_destination.test", &destinationID),
					checkIDEquals("hookdeck_gateway_transformation.test", &transformationID),
					checkIDEquals("hookdeck_gateway_connection.test", &connectionID),
				),
			},
		},
	})
}

// v2 names keep working in v3 as deprecated aliases.
func TestAccV3_V2NamesStillWork(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "v2_names.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hookdeck_source.test", "name", "v2-src-"+suffix),
					resource.TestCheckResourceAttrPair("hookdeck_connection.test", "destination_id", "hookdeck_destination.test", "id"),
				),
			},
		},
	})
}

// The v2 and v3 names are the same resource; a v2 import id works under
// the v3 name.
func TestAccV3_ImportUnderNewName(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "source_without_project_id.tf", suffix),
			},
			{
				ResourceName:            "hookdeck_gateway_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				ResourceName:  "hookdeck_gateway_source.test",
				ImportState:   true,
				ImportStateId: "not/a/valid/id",
				ExpectError:   regexp.MustCompile(`import id must be`),
			},
		},
	})
}
