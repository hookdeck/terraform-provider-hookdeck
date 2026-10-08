package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const releasedV2 = "2.4.0"

var upgradeTypes = []string{"source", "source_auth", "destination", "transformation", "connection"}

// requiredProviders pins the registry address both the released provider
// and the one under test are installed as. An empty version leaves the
// provider under test unconstrained.
func requiredProviders(version string) string {
	constraint := ""
	if version != "" {
		constraint = fmt.Sprintf("\n      version = %q", version)
	}
	return fmt.Sprintf("terraform {\n  required_providers {\n    hookdeck = {\n      source = \"hookdeck/hookdeck\"%s\n    }\n  }\n}\n", constraint)
}

// upgradeSteps starts from state written by the released v2 provider and
// applies upgraded with the provider under test, which must plan nothing,
// keep every id and record each resource's project.
func upgradeSteps(t *testing.T, suffix, upgraded, prefix string) []resource.TestStep {
	t.Helper()
	projectID := currentProjectID(t)
	var capture, compare []resource.TestCheckFunc
	for _, name := range upgradeTypes {
		compare = append(compare, resource.TestCheckResourceAttr(prefix+name+".test", "project_id", projectID))
		if name == "source_auth" {
			continue
		}
		id := new(string)
		capture = append(capture, captureID("hookdeck_"+name+".test", id))
		compare = append(compare, checkIDEquals(prefix+name+".test", id))
	}
	return []resource.TestStep{
		{
			ExternalProviders: map[string]resource.ExternalProvider{
				"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV2},
			},
			Config: requiredProviders(releasedV2) + loadFixture(t, "v2_names.tf", suffix),
			Check:  resource.ComposeAggregateTestCheckFunc(capture...),
		},
		{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Config:                   requiredProviders("") + loadFixture(t, upgraded, suffix),
			ConfigPlanChecks:         emptyPlan(),
			Check:                    resource.ComposeAggregateTestCheckFunc(compare...),
		},
	}
}

// A v2 configuration applied with the released v2 provider keeps working
// unchanged: the plan is empty and each resource is pinned to its project.
func TestAccUpgrade_ReleasedV2State_SameNames(t *testing.T) {
	skipUnlessAcc(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps:    upgradeSteps(t, suffix, "v2_names.tf", "hookdeck_"),
	})
}

// The same state renamed with moved blocks: nothing is recreated.
func TestAccUpgrade_ReleasedV2State_Moved(t *testing.T) {
	skipUnlessAcc(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:               func() { testAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		Steps:                  upgradeSteps(t, suffix, "v3_names_moved.tf", "hookdeck_gateway_"),
	})
}

// State written by the released v2 provider with a project key, taken over
// by an organization key with project_id on each resource. The project in
// the v2 state pins the source, so nothing is planned for it. The v2 state
// of a source auth records no project: it is adopted with one in-place
// update.
func TestAccMock_UpgradeReleasedV2StateToOrgKey(t *testing.T) {
	skipUnlessAcc(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	m := mockOrganization(t)
	useAPIKey(t, mockProjectKeyA)
	var sourceID string

	const (
		v2Source = "hookdeck_source.test"
		v2Auth   = "hookdeck_source_auth.test"
	)
	config := func(projectID string) string {
		return loadFixture(t, "v2_source_with_auth.tf", projectID)
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV2},
				},
				Config: requiredProviders(releasedV2) + config(""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(v2Source, &sourceID),
					resource.TestCheckNoResourceAttr(v2Source, "project_id"),
				),
			},
			{
				PreConfig:                func() { useAPIKey(t, mockOrgKey) },
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   requiredProviders("") + config("  project_id = \"tm_a\"\n"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(v2Source, plancheck.ResourceActionNoop),
					plancheck.ExpectResourceAction(v2Auth, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(v2Source, &sourceID),
					resource.TestCheckResourceAttr(v2Source, "project_id", "tm_a"),
					resource.TestCheckResourceAttr(v2Auth, "project_id", "tm_a"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   requiredProviders("") + config("  project_id = \"tm_a\"\n"),
				ConfigPlanChecks:         emptyPlan(),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   requiredProviders("") + config("  project_id = \"tm_b\"\n"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(v2Source, plancheck.ResourceActionReplace),
					plancheck.ExpectResourceAction(v2Auth, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(v2Source, &sourceID),
					resource.TestCheckResourceAttr(v2Auth, "project_id", "tm_b"),
				),
			},
		},
	})

	for _, r := range m.requestsTo("/sources") {
		if r.key == mockOrgKey && r.project == "" {
			t.Errorf("organization key request without a project: %+v", r)
		}
	}
}

// Released v2 state moved to the v3 names: the v3 names have no team_id,
// and the project it recorded is kept in project_id.
func TestAccMock_UpgradeReleasedV2State_MovedDropsTeamID(t *testing.T) {
	skipUnlessAcc(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	mockOrganization(t)
	useAPIKey(t, mockProjectKeyA)
	var sourceID string

	const (
		v2Source = "hookdeck_source.test"
		v3Source = "hookdeck_gateway_source.test"
		v3Auth   = "hookdeck_gateway_source_auth.test"
	)
	moved := `resource "hookdeck_gateway_source" "test" {
  name = "v2-src"
}

resource "hookdeck_gateway_source_auth" "test" {
  source_id = hookdeck_gateway_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "secret"
  })
}

moved {
  from = hookdeck_source.test
  to   = hookdeck_gateway_source.test
}

moved {
  from = hookdeck_source_auth.test
  to   = hookdeck_gateway_source_auth.test
}
`

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV2},
				},
				Config: requiredProviders(releasedV2) + loadFixture(t, "v2_source_with_auth.tf", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(v2Source, &sourceID),
					resource.TestCheckResourceAttr(v2Source, "team_id", "tm_a"),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Config:                   requiredProviders("") + moved,
				ConfigPlanChecks:         emptyPlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(v3Source, &sourceID),
					resource.TestCheckResourceAttr(v3Source, "project_id", "tm_a"),
					resource.TestCheckNoResourceAttr(v3Source, "team_id"),
					resource.TestCheckResourceAttr(v3Auth, "project_id", "tm_a"),
				),
			},
		},
	})
}
