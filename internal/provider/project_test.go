package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const projectAddr = "hookdeck_gateway_project.test"

func projectConfig(name string) string {
	return fmt.Sprintf("resource \"hookdeck_gateway_project\" \"test\" {\n  name = %q\n}\n", name)
}

// A project key cannot manage projects; the plan says so.
func TestAccGatewayProject_ProjectKeyIsPlanError(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      projectConfig("tf-v3-" + suffix),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Organization API key required`),
			},
		},
	})
}

// A name the API would trim fails validation. The key is never sent, so
// this runs without organization credentials.
func TestAccGatewayProject_Validation(t *testing.T) {
	skipUnlessAcc(t)
	useAPIKey(t, fakeOrgAPIKey)

	var steps []resource.TestStep
	for _, name := range []string{" padded", "padded ", strings.Repeat("x", 41)} {
		steps = append(steps, resource.TestStep{
			Config:      projectConfig(name),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`must not start or end with whitespace|string length must be between 1 and 40`),
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// A project key sees its own project through the data source.
func TestAccGatewayProject_ProjectKey_DataSource(t *testing.T) {
	skipUnlessAcc(t)
	projectID := currentProjectID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "project_datasource.tf", projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_id", "id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_id", "type", "event_gateway"),
					resource.TestCheckResourceAttrSet("data.hookdeck_gateway_project.by_id", "organization_id"),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_name", "id", projectID),
				),
			},
			{
				Config:      loadFixture(t, "project_datasource.tf", missingProjectID),
				ExpectError: regexp.MustCompile(`Project not found`),
			},
		},
	})
}

// Create a project and a source in it, rename the project in place, import
// it, then destroy both.
func TestAccGatewayProject_OrgKey_Lifecycle(t *testing.T) {
	skipUnlessOrg(t)
	orgKey := os.Getenv(envOrgAPIKey)
	useAPIKey(t, orgKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var projectID, sourceID string

	config := func(name string) string {
		return projectConfig(name) + loadFixture(t, "source_in_project.tf", suffix)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if projectID == "" {
				return nil
			}
			status, _ := apiRequest(t, orgKey, "", "GET", "/projects/"+projectID, nil)
			if status != 404 {
				return fmt.Errorf("project %s still exists (status %d)", projectID, status)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config("tf-v3-" + suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(projectAddr, &projectID),
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(projectAddr, "name", "tf-v3-"+suffix),
					resource.TestCheckResourceAttr(projectAddr, "type", "event_gateway"),
					resource.TestCheckResourceAttrSet(projectAddr, "organization_id"),
					resource.TestCheckResourceAttrPair(sourceAddr, "project_id", projectAddr, "id"),
					resource.TestCheckResourceAttrPair(sourceAddr, "team_id", projectAddr, "id"),
				),
			},
			{
				Config: config("tf-v3-" + suffix + "-renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(projectAddr, plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(projectAddr, &projectID),
					checkIDEquals(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(projectAddr, "name", "tf-v3-"+suffix+"-renamed"),
					resource.TestCheckResourceAttrPair(sourceAddr, "project_id", projectAddr, "id"),
				),
			},
			{
				ResourceName:      projectAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGatewayProject_OrgKey_DataSource(t *testing.T) {
	skipUnlessOrg(t)
	useAPIKey(t, os.Getenv(envOrgAPIKey))
	projectID := os.Getenv(envOrgProjectID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "project_datasource.tf", projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_id", "id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_name", "id", projectID),
					resource.TestCheckResourceAttrPair("data.hookdeck_gateway_project.by_name", "name", "data.hookdeck_gateway_project.by_id", "name"),
				),
			},
		},
	})
}
