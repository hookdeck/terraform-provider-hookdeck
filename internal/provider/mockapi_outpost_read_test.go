package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func deleteOutpostBehindTerraform(m *mockAPI, id *string) func() {
	return func() {
		m.update(func() {
			delete(m.projects, *id)
			delete(m.outposts, *id)
		})
	}
}

func outpostKeepsState(t *testing.T, m *mockAPI, otherKey string) {
	t.Helper()
	var projectID string
	patches, puts := 0, 0

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: outpostConfig("deliveries", outpostBasic),
				Check:  captureID(outpostAddr, &projectID),
			},
			{
				PreConfig:   func() { acc.UseAPIKey(t, otherKey) },
				Config:      outpostConfig("deliveries", outpostBasic),
				ExpectError: cannotAccessProject,
			},
			{
				PreConfig: func() {
					acc.UseAPIKey(t, mockOrgKey)
					patches = len(m.sentPatches())
					puts = countRequests(m, http.MethodPut, "/projects/")
				},
				Config:           outpostConfig("deliveries", outpostBasic),
				ConfigPlanChecks: emptyPlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(outpostAddr, &projectID),
					resource.TestCheckResourceAttr(outpostAddr, "config.TOPICS", "a,b"),
					resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "k=v"),
					checkNoNewWrites(m, &patches, &puts),
				),
			},
		},
	})
	if got := countRequests(m, http.MethodPost, "/projects"); got != 1 {
		t.Errorf("%d POST /projects requests, want 1", got)
	}
}

// A key of another organization gets 401 for the project: error, project and
// its configuration kept, no second project created.
func TestAccMock_OutpostProject_OtherOrganizationKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addKey(mockOtherOrgKey, mockKey{org: mockOtherOrg})
	outpostKeepsState(t, m, mockOtherOrgKey)
}

// A key limited to other projects gets 403 for the project: error, project
// kept.
func TestAccMock_OutpostProject_UngrantedKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addKey(mockGrantedKey, mockKey{org: mockOrg, grants: map[string]bool{"tm_a": true}})
	outpostKeepsState(t, m, mockGrantedKey)
}

// Destroying, or removing the block, when the project is already gone
// succeeds.
func TestAccMock_OutpostProject_DestroyAlreadyDeleted(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: outpostConfig("deliveries", outpostBasic),
				Check:  captureID(outpostAddr, &projectID),
			},
			{
				PreConfig: deleteOutpostBehindTerraform(m, &projectID),
				Config:    "# no resources\n",
			},
			{
				Config: outpostConfig("receipts", outpostBasic),
				Check:  captureID(outpostAddr, &projectID),
			},
			{
				PreConfig: deleteOutpostBehindTerraform(m, &projectID),
				Config:    outpostConfig("receipts", outpostBasic),
				Destroy:   true,
			},
		},
	})
}

// Importing a project of another organization says the key cannot reach it.
func TestAccMock_OutpostProject_ImportUnreachable(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProjectIn(mockOtherOrg, mockOtherOrgProject, "elsewhere", "outpost")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config:        outpostConfig("deliveries", ""),
			ResourceName:  outpostAddr,
			ImportState:   true,
			ImportStateId: mockOtherOrgProject,
			ExpectError:   cannotAccessProject,
		}},
	})
}

// With an organization key, the data source says the key cannot reach a
// project of another organization; an unknown ID is not found.
func TestAccMock_OutpostProject_DataSourceUnreachable(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProjectIn(mockOtherOrg, mockOtherOrgProject, "elsewhere", "outpost")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      "data \"hookdeck_outpost_project\" \"test\" {\n  id = \"" + mockOtherOrgProject + "\"\n}\n",
				ExpectError: cannotAccessProject,
			},
			{
				Config:      "data \"hookdeck_outpost_project\" \"test\" {\n  id = \"tm_doesnotexist123\"\n}\n",
				ExpectError: regexp.MustCompile(`Project not found`),
			},
		},
	})
}
