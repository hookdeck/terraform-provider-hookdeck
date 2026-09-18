package provider_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// --- project API key ---

func TestAccV3_ProjectKey_ProjectIDMatchesKey(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "source_with_project_id.tf", suffix, currentProjectID(t)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "project_id", currentProjectID(t)),
					resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "team_id", currentProjectID(t)),
				),
			},
			{
				ResourceName:            "hookdeck_gateway_source.test",
				ImportState:             true,
				ImportStateIdFunc:       importIDWithProject("hookdeck_gateway_source.test", currentProjectID(t)),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

func TestAccV3_ProjectKey_ProjectIDMismatch(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      loadFixture(t, "source_with_project_id.tf", suffix, "tm_doesnotexist"),
				ExpectError: regexp.MustCompile(`project_id does not match`),
			},
		},
	})
}

func TestAccV3_ProjectKey_ProviderProjectIDMismatch(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      loadFixture(t, "provider_project_id.tf", suffix, "tm_doesnotexist"),
				ExpectError: regexp.MustCompile(`project_id does not match`),
			},
		},
	})
}

func TestAccV3_ProjectKey_ProviderProjectIDMatchesKey(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "provider_project_id.tf", suffix, currentProjectID(t)),
				Check:  resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "team_id", currentProjectID(t)),
			},
		},
	})
}

// --- organization API key ---

func TestAccV3_OrgKey_ProjectRequired(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccOrgPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      orgProvider("") + loadFixture(t, "source_without_project_id.tf", suffix),
				ExpectError: regexp.MustCompile(`project_id is required`),
			},
		},
	})
}

func TestAccV3_OrgKey_ProviderDefaultProject(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := os.Getenv(envOrgProjectID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccOrgPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: orgProvider(projectID) + loadFixture(t, "source_without_project_id.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("hookdeck_gateway_source.test", "project_id"),
					resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "team_id", projectID),
				),
			},
			{
				// Bare id import resolves the project from the provider.
				ResourceName:            "hookdeck_gateway_source.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

func TestAccV3_OrgKey_ResourceProjectOverridesProvider(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := os.Getenv(envOrgProjectID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccOrgPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Provider default points at a project that does not exist;
				// the resource's own project_id must win.
				Config: orgProvider("tm_doesnotexist") + loadFixture(t, "source_with_project_id.tf", suffix, projectID),
				Check:  resource.TestCheckResourceAttr("hookdeck_gateway_source.test", "team_id", projectID),
			},
			{
				ResourceName:            "hookdeck_gateway_source.test",
				ImportState:             true,
				ImportStateIdFunc:       importIDWithProject("hookdeck_gateway_source.test", projectID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

// Create a project with the organization key, put a source in it, rename
// the project in place, then destroy both.
func TestAccV3_OrgKey_GatewayProjectLifecycle(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var projectID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccOrgPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			status, _ := apiRequest(t, os.Getenv(envOrgAPIKey), "GET", "/2026-09-01/projects/"+projectID)
			if status != 404 {
				return fmt.Errorf("project %s still exists (status %d)", projectID, status)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: orgProvider("") + loadFixture(t, "org_project_lifecycle.tf", suffix, "tf-v3-"+suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID("hookdeck_gateway_project.test", &projectID),
					resource.TestCheckResourceAttr("hookdeck_gateway_project.test", "name", "tf-v3-"+suffix),
					resource.TestCheckResourceAttr("hookdeck_gateway_project.test", "type", "event_gateway"),
					resource.TestCheckResourceAttrSet("hookdeck_gateway_project.test", "organization_id"),
					resource.TestCheckResourceAttrPair("hookdeck_gateway_source.test", "team_id", "hookdeck_gateway_project.test", "id"),
				),
			},
			{
				Config: orgProvider("") + loadFixture(t, "org_project_lifecycle.tf", suffix, "tf-v3-"+suffix+"-renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals("hookdeck_gateway_project.test", &projectID),
					resource.TestCheckResourceAttr("hookdeck_gateway_project.test", "name", "tf-v3-"+suffix+"-renamed"),
				),
			},
			{
				ResourceName:      "hookdeck_gateway_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccV3_OrgKey_GatewayProjectDataSource(t *testing.T) {
	projectID := os.Getenv(envOrgProjectID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccOrgPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: orgProvider("") + loadFixture(t, "org_project_datasource.tf", projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_id", "id", projectID),
					resource.TestCheckResourceAttr("data.hookdeck_gateway_project.by_name", "id", projectID),
					resource.TestCheckResourceAttrPair("data.hookdeck_gateway_project.by_name", "name", "data.hookdeck_gateway_project.by_id", "name"),
				),
			},
		},
	})
}

func importIDWithProject(name, projectID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", name)
		}
		return projectID + "/" + rs.Primary.ID, nil
	}
}
