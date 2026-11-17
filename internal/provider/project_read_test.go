package provider_test

import (
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const (
	mockOtherOrg        = "org_2"
	mockOtherOrgKey     = "hd_org_mock_other"
	mockOtherOrgProject = "tm_other"
	mockGrantedKey      = "hd_org_mock_granted"
)

var cannotAccessProject = regexp.MustCompile(`(?s)cannot\s+access\s+project.*Use\s+an\s+API\s+key`)

// deleteProjectBehindTerraform removes the project from the mock API, as a
// deletion outside Terraform would.
func deleteProjectBehindTerraform(m *mockAPI, id *string) func() {
	return func() { m.update(func() { delete(m.projects, *id) }) }
}

func recreatePlan() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(projectAddr, plancheck.ResourceActionCreate),
	}}
}

// A project deleted outside Terraform is removed from state and created
// again.
func TestAccMock_GatewayProject_DeletedIsRecreated(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockOrgKey)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: projectConfig("gamma"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig:        deleteProjectBehindTerraform(m, &projectID),
				Config:           projectConfig("gamma"),
				ConfigPlanChecks: recreatePlan(),
				Check:            checkIDChanged(projectAddr, &projectID),
			},
		},
	})
}

// Destroying, or removing the block, when the project is already gone
// succeeds.
func TestAccMock_GatewayProject_DestroyAlreadyDeleted(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockOrgKey)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: projectConfig("gamma"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig: deleteProjectBehindTerraform(m, &projectID),
				Config:    "# no resources\n",
			},
			{
				Config: projectConfig("delta"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig: deleteProjectBehindTerraform(m, &projectID),
				Config:    projectConfig("delta"),
				Destroy:   true,
			},
		},
	})
}

// A key of another organization gets 401 for the project: error, project kept.
func TestAccMock_GatewayProject_OtherOrganizationKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	m.addKey(mockOtherOrgKey, mockKey{org: mockOtherOrg})
	acc.UseAPIKey(t, mockOrgKey)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: projectConfig("gamma"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig:   func() { acc.UseAPIKey(t, mockOtherOrgKey) },
				Config:      projectConfig("gamma"),
				ExpectError: cannotAccessProject,
			},
			{
				PreConfig:        func() { acc.UseAPIKey(t, mockOrgKey) },
				Config:           projectConfig("gamma"),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(projectAddr, &projectID),
			},
		},
	})
}

// A key limited to other projects gets 403 for the project: error, project
// kept.
func TestAccMock_GatewayProject_UngrantedKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	m.addKey(mockGrantedKey, mockKey{org: mockOrg, grants: map[string]bool{"tm_a": true}})
	acc.UseAPIKey(t, mockOrgKey)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: projectConfig("gamma"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig:   func() { acc.UseAPIKey(t, mockGrantedKey) },
				Config:      projectConfig("gamma"),
				ExpectError: cannotAccessProject,
			},
			{
				PreConfig:        func() { acc.UseAPIKey(t, mockOrgKey) },
				Config:           projectConfig("gamma"),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(projectAddr, &projectID),
			},
		},
	})
}

// Importing an unknown ID fails with Terraform's own error; importing a
// project of another organization says the key cannot reach it.
func TestAccMock_GatewayProject_ImportUnreachable(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	m.addProjectIn(mockOtherOrg, mockOtherOrgProject, "elsewhere", "event_gateway")
	acc.UseAPIKey(t, mockOrgKey)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        projectConfig("gamma"),
				ResourceName:  projectAddr,
				ImportState:   true,
				ImportStateId: "tm_doesnotexist123",
				ExpectError:   regexp.MustCompile(`Cannot\s+import\s+non-existent\s+remote\s+object`),
			},
			{
				Config:        projectConfig("gamma"),
				ResourceName:  projectAddr,
				ImportState:   true,
				ImportStateId: mockOtherOrgProject,
				ExpectError:   cannotAccessProject,
			},
		},
	})
}

// A project key reads its own project through the data source; another
// project's ID is not found, as the API answers 404 when no header names it.
func TestAccMock_GatewayProject_DataSourceProjectKey(t *testing.T) {
	acc.SkipUnlessAcc(t)
	mockOrganization(t)
	acc.UseAPIKey(t, mockProjectKeyA)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "data \"hookdeck_gateway_project\" \"test\" {\n  id = \"tm_a\"\n}\n",
				Check:  resource.TestCheckResourceAttr("data.hookdeck_gateway_project.test", "name", "alpha"),
			},
			{
				Config:      "data \"hookdeck_gateway_project\" \"test\" {\n  id = \"tm_b\"\n}\n",
				ExpectError: regexp.MustCompile(`Project not found`),
			},
		},
	})
}
