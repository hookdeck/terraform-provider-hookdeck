package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// Every Event Gateway resource is reachable under its gateway_ name and
// round-trips through import.
func TestAccNaming_GatewayResourceNames(t *testing.T) {
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
					resource.TestCheckResourceAttrPair("hookdeck_gateway_source_auth.test", "project_id", "hookdeck_gateway_source.test", "project_id"),
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
			{
				ResourceName:                         "hookdeck_gateway_source_auth.test",
				ImportState:                          true,
				ImportStateIdFunc:                    importIDFromAttribute("hookdeck_gateway_source_auth.test", "source_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "source_id",
			},
			{
				ResourceName:  "hookdeck_gateway_source.test",
				ImportState:   true,
				ImportStateId: "not/a/valid/id",
				ExpectError:   regexp.MustCompile(`import ID must be`),
			},
			{
				ResourceName:  "hookdeck_gateway_source.test",
				ImportState:   true,
				ImportStateId: "src_doesnotexist",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

// The gateway_ data sources read what the resources created, and record
// the project they read from.
func TestAccNaming_GatewayDataSources(t *testing.T) {
	skipUnlessAcc(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := currentProjectID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "gateway_basic.tf", suffix) + loadFixture(t, "gateway_datasources.tf", projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hookdeck_gateway_source.test", "name", "v3-src-"+suffix),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_source.test", "project_id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_destination.test", "name", "v3-dst-"+suffix),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_destination.test", "project_id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_connection.test", "name", "v3-con-"+suffix),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_connection.test", "project_id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_source.legacy", "name", "v3-src-"+suffix),
				),
			},
			{
				Config:      loadFixture(t, "source_datasource.tf", "src_doesnotexist", ""),
				ExpectError: regexp.MustCompile(`Source not found`),
			},
		},
	})
}

// v2 names keep working in v3 as deprecated aliases.
func TestAccNaming_V2NamesStillWork(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "v2_names.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hookdeck_source.test", "name", "v2-src-"+suffix),
					resource.TestCheckResourceAttr("hookdeck_source_auth.test", "auth_type", "API_KEY"),
					resource.TestCheckResourceAttrPair("hookdeck_connection.test", "destination_id", "hookdeck_destination.test", "id"),
				),
			},
		},
	})
}

// State written under the v2 aliases moves to the gateway_ names with moved
// blocks: nothing is recreated.
func TestAccNaming_MovedFromV2Names(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var capture, compare []resource.TestCheckFunc
	for _, name := range []string{"source", "destination", "transformation", "connection"} {
		id := new(string)
		capture = append(capture, captureID("hookdeck_"+name+".test", id))
		compare = append(compare, checkIDEquals("hookdeck_gateway_"+name+".test", id))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "v2_names.tf", suffix),
				Check:  resource.ComposeAggregateTestCheckFunc(capture...),
			},
			{
				Config:           loadFixture(t, "v3_names_moved.tf", suffix),
				ConfigPlanChecks: emptyPlan(),
				Check: resource.ComposeAggregateTestCheckFunc(append(compare,
					resource.TestCheckResourceAttrPair("hookdeck_gateway_source_auth.test", "source_id", "hookdeck_gateway_source.test", "id"),
				)...),
			},
		},
	})
}
