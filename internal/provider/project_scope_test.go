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

const sourceAddr = "hookdeck_gateway_source.test"

func emptyPlan() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}
}

func replaceSourcePlan() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionReplace),
	}}
}

// checkSourceGone fails unless the source is gone from projectID.
func checkSourceGone(t *testing.T, apiKey, projectID string, sourceID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		status, _ := apiRequest(t, apiKey, projectID, "GET", "/sources/"+*sourceID, nil)
		if status != 404 && status != 410 {
			return fmt.Errorf("source %s still exists in project %s (status %d)", *sourceID, projectID, status)
		}
		return nil
	}
}

// --- project API key: single-project mode ---

// Every resource records its project. Setting project_id to the key's own
// project, or removing it again, changes nothing.
func TestAccProjectScope_ProjectKey_StoresProject(t *testing.T) {
	skipUnlessAcc(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := currentProjectID(t)
	var sourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, suffix, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectID),
					resource.TestCheckResourceAttr(sourceAddr, "team_id", projectID),
				),
			},
			{
				Config:           sourceConfig(t, suffix, projectID),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
			{
				Config:           sourceConfig(t, suffix, ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
			{
				ResourceName:            sourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				ResourceName:            sourceAddr,
				ImportState:             true,
				ImportStateIdFunc:       importSourceWithProject(projectID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				ResourceName:      sourceAddr,
				ImportState:       true,
				ImportStateIdFunc: importSourceWithProject(missingProjectID),
				ExpectError:       regexp.MustCompile(`Project mismatch`),
			},
		},
	})
}

// Importing <project_id>/<id> into a configuration without project_id
// adopts the resource; it is not replaced.
func TestAccProjectScope_ProjectKey_ImportWithProjectKeepsResource(t *testing.T) {
	skipUnlessAcc(t)
	testAccPreCheck(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := currentProjectID(t)

	status, body := apiRequest(t, os.Getenv(envAPIKey), "", "POST", "/sources", map[string]any{"name": "v3-src-" + suffix, "type": "HTTP"})
	if status > 299 {
		t.Fatalf("creating source: %d %v", status, body)
	}
	sourceID, _ := body["id"].(string)
	t.Cleanup(func() { apiRequest(t, os.Getenv(envAPIKey), "", "DELETE", "/sources/"+sourceID, nil) })

	importBlock := fmt.Sprintf("import {\n  to = %s\n  id = \"%s/%s\"\n}\n", sourceAddr, projectID, sourceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: importBlock + sourceConfig(t, suffix, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectID),
				),
			},
		},
	})
}

// A project_id other than the key's project is refused at plan time, for a
// new resource and for an existing one, which is left alone.
func TestAccProjectScope_ProjectKey_OtherProjectIsPlanError(t *testing.T) {
	skipUnlessAcc(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	var sourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      sourceConfig(t, suffix, missingProjectID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Project mismatch`),
			},
			{
				Config: sourceConfig(t, suffix, ""),
				Check:  captureID(sourceAddr, &sourceID),
			},
			{
				Config:      sourceConfig(t, suffix, missingProjectID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Project mismatch(.|\n)*` + missingProjectID + `(.|\n)*` + currentProjectID(t)),
			},
			{
				Config:           sourceConfig(t, suffix, ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}

func TestAccProjectScope_ProjectKey_EmptyProjectIDIsInvalid(t *testing.T) {
	skipUnlessAcc(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      loadFixture(t, "source_with_project_id.tf", suffix, ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)project_id.*string length must be at least 1`),
			},
		},
	})
}

// A provider project_id equal to the key's project works without a lookup.
// One the key cannot reach is refused by the API for a new resource, and
// for an existing resource the plan fails naming both projects; the
// resource stays in state.
func TestAccProjectScope_ProjectKey_ProviderProjectID(t *testing.T) {
	skipUnlessAcc(t)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectID := currentProjectID(t)
	var sourceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerBlock(missingProjectID) + sourceConfig(t, suffix, ""),
				ExpectError: regexp.MustCompile(`cannot access project ` + missingProjectID),
			},
			{
				Config: providerBlock(projectID) + sourceConfig(t, suffix, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectID),
				),
			},
			{
				Config:      providerBlock(missingProjectID) + sourceConfig(t, suffix, ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Project not reachable.*` + projectID + `.*` + missingProjectID + `.*state rm`),
			},
			{
				Config:           sourceConfig(t, suffix, ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}

// --- organization API key ---

// Explicit mode without project_id fails at plan time. The key is never
// sent, so this runs without organization credentials.
func TestAccProjectScope_OrgKey_ProjectRequired(t *testing.T) {
	skipUnlessAcc(t)
	useAPIKey(t, fakeOrgAPIKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      sourceConfig(t, suffix, ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing project_id`),
			},
			{
				Config:      loadFixture(t, "source_datasource.tf", "src_missing", ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing project_id`),
			},
		},
	})
}

// Explicit mode: resources in two projects under one provider. Changing a
// resource's project_id replaces it. Removing project_id or pointing it at
// a project the key cannot see fails the plan and leaves state alone.
func TestAccProjectScope_OrgKey_ExplicitMode(t *testing.T) {
	skipUnlessOrg(t)
	orgKey := os.Getenv(envOrgAPIKey)
	useAPIKey(t, orgKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectA := os.Getenv(envOrgProjectID)
	projectB := secondOrgProject(t)
	var sourceID, movedID string

	const otherAddr = "hookdeck_gateway_source.other"
	config := func(testProject string) string {
		return sourceConfig(t, suffix, testProject) + loadFixture(t, "source_other.tf", suffix, projectB)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(projectA),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectA),
					resource.TestCheckResourceAttr(sourceAddr, "team_id", projectA),
					resource.TestCheckResourceAttr(otherAddr, "project_id", projectB),
					resource.TestCheckResourceAttr(otherAddr, "team_id", projectB),
				),
			},
			{
				ResourceName:            sourceAddr,
				ImportState:             true,
				ImportStateIdFunc:       importSourceWithProject(projectA),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				ResourceName: sourceAddr,
				ImportState:  true,
				ExpectError:  regexp.MustCompile(`Missing project in import ID`),
			},
			{
				Config:      config(""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing project_id`),
			},
			{
				Config:      config(missingProjectID),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`cannot access project ` + missingProjectID),
			},
			{
				Config:           config(projectA),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
			{
				Config:           config(projectB),
				ConfigPlanChecks: replaceSourcePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(sourceAddr, &sourceID),
					captureID(sourceAddr, &movedID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectB),
					resource.TestCheckResourceAttr(sourceAddr, "team_id", projectB),
					checkSourceGone(t, orgKey, projectA, &sourceID),
				),
			},
		},
	})
}

// Single-project mode with an organization key: the provider project_id
// applies to every resource, a resource cannot name another project, and
// changing the provider project_id replaces the resource. A provider
// project_id the key cannot see fails the plan and leaves state alone.
func TestAccProjectScope_OrgKey_SingleProjectMode(t *testing.T) {
	skipUnlessOrg(t)
	orgKey := os.Getenv(envOrgAPIKey)
	useAPIKey(t, orgKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectA := os.Getenv(envOrgProjectID)
	projectB := secondOrgProject(t)
	var sourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerBlock(projectA) + sourceConfig(t, suffix, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectA),
				),
			},
			{
				ResourceName:            sourceAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
			{
				Config:           providerBlock(projectA) + sourceConfig(t, suffix, projectA),
				ConfigPlanChecks: emptyPlan(),
			},
			{
				Config:      providerBlock(projectA) + sourceConfig(t, suffix, projectB),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Project mismatch`),
			},
			{
				Config:      providerBlock(missingProjectID) + sourceConfig(t, suffix, ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`cannot access project ` + missingProjectID),
			},
			{
				Config:           providerBlock(projectA) + sourceConfig(t, suffix, ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
			{
				Config:           providerBlock(projectB) + sourceConfig(t, suffix, ""),
				ConfigPlanChecks: replaceSourcePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", projectB),
					checkSourceGone(t, orgKey, projectA, &sourceID),
				),
			},
		},
	})
}

// A source and its auth move together when the project changes, and the
// new source ends up with the auth set.
func TestAccProjectScope_OrgKey_SourceAuthFollowsSource(t *testing.T) {
	skipUnlessOrg(t)
	orgKey := os.Getenv(envOrgAPIKey)
	useAPIKey(t, orgKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	projectA := os.Getenv(envOrgProjectID)
	projectB := secondOrgProject(t)
	var sourceID string

	const authAddr = "hookdeck_gateway_source_auth.test"
	checkAuth := func(projectID string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			id := s.RootModule().Resources[sourceAddr].Primary.ID
			status, body := apiRequest(t, orgKey, projectID, "GET", "/sources/"+id, nil)
			config, _ := body["config"].(map[string]interface{})
			if status != 200 || config["auth_type"] != "API_KEY" {
				return fmt.Errorf("source %s in %s: status %d, auth_type %v", id, projectID, status, config["auth_type"])
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerBlock(projectA) + loadFixture(t, "source_with_auth.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(authAddr, "project_id", projectA),
					checkAuth(projectA),
				),
			},
			{
				Config: providerBlock(projectB) + loadFixture(t, "source_with_auth.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionReplace),
					plancheck.ExpectResourceAction(authAddr, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(authAddr, "project_id", projectB),
					checkAuth(projectB),
				),
			},
		},
	})
}

// A resource created with an organization key is in a project the project
// key cannot reach. With the project key the plan fails naming both
// projects, and the resource is still in state afterwards.
func TestAccProjectScope_KeyCannotReachStoredProject(t *testing.T) {
	skipUnlessOrg(t)
	testAccPreCheck(t)
	projectKey := os.Getenv(envAPIKey)
	orgKey := os.Getenv(envOrgAPIKey)
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	orgProject := os.Getenv(envOrgProjectID)
	keyProject := currentProjectID(t)
	var sourceID string

	useAPIKey(t, orgKey)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, suffix, orgProject),
				Check:  captureID(sourceAddr, &sourceID),
			},
			{
				PreConfig:   func() { useAPIKey(t, projectKey) },
				Config:      sourceConfig(t, suffix, ""),
				ExpectError: regexp.MustCompile(`(?s)Project not reachable.*` + orgProject + `.*` + keyProject + `.*state rm`),
			},
			{
				PreConfig:        func() { useAPIKey(t, orgKey) },
				Config:           sourceConfig(t, suffix, orgProject),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}
