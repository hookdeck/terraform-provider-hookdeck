package provider_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

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
		PreCheck:                 func() { acc.PreCheckProjectKey(t) },
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      projectConfig("tf-acc-v3-" + suffix),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Organization API key required`),
			},
		},
	})
}

// A name the API would trim fails validation. The key is never sent, so
// this runs without organization credentials.
func TestAccGatewayProject_Validation(t *testing.T) {
	acc.SkipUnlessAcc(t)
	acc.UseAPIKey(t, acc.FakeOrgAPIKey)

	var steps []resource.TestStep
	for _, name := range []string{" padded", "padded ", strings.Repeat("x", 41)} {
		steps = append(steps, resource.TestStep{
			Config:      projectConfig(name),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`must not start or end with whitespace|string length must be between 1 and 40`),
		})
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps:                    steps,
	})
}

// A project key sees its own project through the data source.
func TestAccGatewayProject_ProjectKey_DataSource(t *testing.T) {
	acc.SkipUnlessAcc(t)
	projectID := acc.ProjectA(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acc.PreCheckProjectKey(t) },
		ProtoV6ProviderFactories: acc.ProviderFactories,
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
				Config:      loadFixture(t, "project_datasource.tf", acc.MissingProjectID),
				ExpectError: regexp.MustCompile(`Project not found`),
			},
		},
	})
}

// Create a project and a source in it, rename the project in place, import
// it, then destroy both.
func TestAccGatewayProject_OrgKey_Lifecycle(t *testing.T) {
	acc.PreCheckOrgKey(t)
	orgKey := acc.OrgKey(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var projectID, sourceID string

	config := func(name string) string {
		return projectConfig(name) + loadFixture(t, "source_in_project.tf", suffix)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			if projectID == "" {
				return nil
			}
			status, _ := acc.Request(t, orgKey, "", "GET", "/projects/"+projectID, nil)
			if status != 404 {
				return fmt.Errorf("project %s still exists (status %d)", projectID, status)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config("tf-acc-v3-" + suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(projectAddr, &projectID),
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(projectAddr, "name", "tf-acc-v3-"+suffix),
					resource.TestCheckResourceAttr(projectAddr, "type", "event_gateway"),
					resource.TestCheckResourceAttrSet(projectAddr, "organization_id"),
					resource.TestCheckResourceAttrPair(sourceAddr, "project_id", projectAddr, "id"),
				),
			},
			{
				Config: config("tf-acc-v3-" + suffix + "-renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(projectAddr, plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(projectAddr, &projectID),
					checkIDEquals(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(projectAddr, "name", "tf-acc-v3-"+suffix+"-renamed"),
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
	acc.PreCheckOrgKey(t)
	projectID := acc.ProjectA(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
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
