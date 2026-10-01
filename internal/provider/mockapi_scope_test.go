package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	mockOrgKey      = "hd_org_mock"
	mockProjectKeyA = "hd_mock_a"
	mockProjectKeyB = "hd_mock_b"
)

// mockOrganization has two Event Gateway projects, a project key for each
// and an organization key.
func mockOrganization(t *testing.T) *mockAPI {
	t.Helper()
	m := newMockAPI(t)
	m.addProject("tm_a", "alpha", "event_gateway")
	m.addProject("tm_b", "beta", "event_gateway")
	m.addKey(mockProjectKeyA, mockKey{org: mockOrg, project: "tm_a"})
	m.addKey(mockProjectKeyB, mockKey{org: mockOrg, project: "tm_b"})
	m.addKey(mockOrgKey, mockKey{org: mockOrg})
	return m
}

// A project key that may not list projects still works: the project is
// taken from the API's responses. Swapping in another project's key is
// refused by the API and reported as a project problem, not as a missing
// resource.
func TestAccMock_ProjectKeyWithoutProjectsRead(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	m.update(func() {
		m.keys[mockProjectKeyA].noProjectsRead = true
		m.keys[mockProjectKeyB].noProjectsRead = true
	})
	useAPIKey(t, mockProjectKeyA)
	var sourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, "mock", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", "tm_a"),
				),
			},
			{
				Config:           sourceConfig(t, "mock", "tm_a"),
				ConfigPlanChecks: emptyPlan(),
			},
			{
				Config:      sourceConfig(t, "mock", "tm_b"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Project not reachable.*tm_a.*tm_b`),
			},
			{
				PreConfig:   func() { useAPIKey(t, mockProjectKeyB) },
				Config:      sourceConfig(t, "mock", ""),
				ExpectError: regexp.MustCompile(`(?s)cannot access project tm_a.*state rm`),
			},
			{
				PreConfig:        func() { useAPIKey(t, mockProjectKeyA) },
				Config:           sourceConfig(t, "mock", ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}

// Swapping the project key for another project's key fails the plan before
// any request for the resource is sent, and leaves the state alone.
func TestAccMock_ProjectKeySwapped(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	useAPIKey(t, mockProjectKeyA)
	var sourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, "mock", ""),
				Check:  captureID(sourceAddr, &sourceID),
			},
			{
				PreConfig:   func() { useAPIKey(t, mockProjectKeyB) },
				Config:      sourceConfig(t, "mock", ""),
				ExpectError: regexp.MustCompile(`(?s)Project not reachable.*tm_a.*tm_b.*state rm`),
			},
			{
				PreConfig:        func() { useAPIKey(t, mockProjectKeyA) },
				Config:           sourceConfig(t, "mock", ""),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})

	for _, r := range m.requestsTo("/sources") {
		if r.key == mockProjectKeyB {
			t.Errorf("request sent with the swapped key: %+v", r)
		}
	}
}

// An organization key that loses its grant on a project gets an error
// naming the project. The resource is not dropped from state.
func TestAccMock_OrgKeyLosesGrant(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	m.update(func() { m.keys[mockOrgKey].grants = map[string]bool{"tm_a": true} })
	useAPIKey(t, mockOrgKey)
	var sourceID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, "mock", "tm_a"),
				Check:  captureID(sourceAddr, &sourceID),
			},
			{
				PreConfig:   func() { m.update(func() { m.keys[mockOrgKey].grants = map[string]bool{} }) },
				Config:      sourceConfig(t, "mock", "tm_a"),
				ExpectError: regexp.MustCompile(`(?s)cannot access project tm_a.*no grant`),
			},
			{
				PreConfig:        func() { m.update(func() { m.keys[mockOrgKey].grants = map[string]bool{"tm_a": true} }) },
				Config:           sourceConfig(t, "mock", "tm_a"),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}

// A project deleted outside Terraform, or a project_id that never existed,
// is an error on refresh. The resources in it are not dropped from state.
func TestAccMock_OrgKeyProjectNotFoundKeepsState(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	useAPIKey(t, mockOrgKey)
	var sourceID string
	var removed *mockProject

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: sourceConfig(t, "mock", "tm_a"),
				Check:  captureID(sourceAddr, &sourceID),
			},
			{
				PreConfig: func() {
					m.update(func() {
						removed = m.projects["tm_a"]
						delete(m.projects, "tm_a")
					})
				},
				Config:      sourceConfig(t, "mock", "tm_a"),
				ExpectError: regexp.MustCompile(`(?s)cannot access project tm_a.*does not exist`),
			},
			{
				PreConfig:        func() { m.update(func() { m.projects["tm_a"] = removed }) },
				Config:           sourceConfig(t, "mock", "tm_a"),
				ConfigPlanChecks: emptyPlan(),
				Check:            checkIDEquals(sourceAddr, &sourceID),
			},
		},
	})
}

// The two modes of an organization key, end to end: explicit mode needs
// project_id on the resource and replaces on a change; single-project mode
// refuses another project on a resource and replaces on a provider change.
// A target project that does not exist fails the plan before anything is
// deleted.
func TestAccMock_OrgKeyModes(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	useAPIKey(t, mockOrgKey)
	var sourceID string
	deletes := func() (n int) {
		for _, r := range m.requestsTo("/sources") {
			if r.method == "DELETE" {
				n++
			}
		}
		return n
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      sourceConfig(t, "mock", ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Missing project_id`),
			},
			{
				Config: sourceConfig(t, "mock", "tm_a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", "tm_a"),
				),
			},
			{
				Config:      sourceConfig(t, "mock", "tm_missing"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`cannot access project tm_missing`),
			},
			{
				Config:           sourceConfig(t, "mock", "tm_b"),
				ConfigPlanChecks: replaceSourcePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(sourceAddr, &sourceID),
					captureID(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", "tm_b"),
				),
			},
			{
				Config:           providerBlock("tm_b") + sourceConfig(t, "mock", ""),
				ConfigPlanChecks: emptyPlan(),
			},
			{
				Config:      providerBlock("tm_b") + sourceConfig(t, "mock", "tm_a"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Project mismatch`),
			},
			{
				PreConfig: func() {
					if got := deletes(); got != 1 {
						t.Errorf("%d deletes before the provider project changes, want 1", got)
					}
				},
				Config:      providerBlock("tm_missing") + sourceConfig(t, "mock", ""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`cannot access project tm_missing`),
			},
			{
				Config:           providerBlock("tm_a") + sourceConfig(t, "mock", ""),
				ConfigPlanChecks: replaceSourcePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDChanged(sourceAddr, &sourceID),
					resource.TestCheckResourceAttr(sourceAddr, "project_id", "tm_a"),
				),
			},
		},
	})
}

// With an organization key every request for a resource names the
// resource's project, and the provider looks nothing up at configure time.
func TestAccMock_OrgKeySendsTheProjectOnEveryRequest(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	useAPIKey(t, mockOrgKey)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: sourceConfig(t, "mock", "tm_a")},
			{Config: sourceConfig(t, "mock-renamed", "tm_a")},
			{
				ResourceName:            sourceAddr,
				ImportState:             true,
				ImportStateIdFunc:       importSourceWithProject("tm_a"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})

	methods := map[string]bool{}
	for _, r := range m.requestsTo("/sources") {
		methods[r.method] = true
		if r.project != "tm_a" {
			t.Errorf("%s %s sent with project %q, want tm_a", r.method, r.path, r.project)
		}
	}
	for _, method := range []string{"POST", "GET", "PUT", "DELETE"} {
		if !methods[method] {
			t.Errorf("no %s request for the source was recorded", method)
		}
	}
	if lookups := m.requestsTo("/projects"); len(lookups) != 0 {
		t.Errorf("unexpected project requests: %+v", lookups)
	}
}

// A project key looks its project up at configure time, unless project_id
// is given.
func TestAccMock_ProjectKeyLookup(t *testing.T) {
	skipUnlessAcc(t)

	t.Run("without project_id", func(t *testing.T) {
		m := mockOrganization(t)
		useAPIKey(t, mockProjectKeyA)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps:                    []resource.TestStep{{Config: sourceConfig(t, "mock", "")}},
		})
		if len(m.requestsTo("/projects")) == 0 {
			t.Error("the key's project was never looked up")
		}
		for _, r := range m.requestsTo("/sources") {
			if r.project != "tm_a" {
				t.Errorf("%s %s sent with project %q, want tm_a", r.method, r.path, r.project)
			}
		}
	})

	t.Run("with a provider project_id", func(t *testing.T) {
		m := mockOrganization(t)
		useAPIKey(t, mockProjectKeyA)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps:                    []resource.TestStep{{Config: providerBlock("tm_a") + sourceConfig(t, "mock", "")}},
		})
		if lookups := m.requestsTo("/projects"); len(lookups) != 0 {
			t.Errorf("unexpected project requests: %+v", lookups)
		}
	})

	t.Run("with the environment variable", func(t *testing.T) {
		m := mockOrganization(t)
		useAPIKey(t, mockOrgKey)
		t.Setenv(envProjectID, "tm_b")
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config: sourceConfig(t, "mock", ""),
				Check:  resource.TestCheckResourceAttr(sourceAddr, "project_id", "tm_b"),
			}},
		})
		if len(m.requestsTo("/sources")) == 0 {
			t.Error("no source requests recorded")
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		mockOrganization(t)
		useAPIKey(t, "hd_mock_unknown")
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config:      sourceConfig(t, "mock", ""),
				ExpectError: regexp.MustCompile(`Unable to look up the API key's project`),
			}},
		})
	})
}

// hookdeck_gateway_project only accepts Event Gateway projects, on import
// and in the data source, by ID and by name.
func TestAccMock_GatewayProject_OtherProjectTypes(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	m.addProject("tm_out", "deliveries", "outpost")
	m.addProject("tm_out2", "alpha", "outpost")
	useAPIKey(t, mockOrgKey)

	byName := func(name string) string {
		return "data \"hookdeck_gateway_project\" \"test\" {\n  name = \"" + name + "\"\n}\n"
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        projectConfig("deliveries"),
				ResourceName:  projectAddr,
				ImportState:   true,
				ImportStateId: "tm_out",
				ExpectError:   regexp.MustCompile(`(?s)Unsupported project type.*outpost`),
			},
			{
				Config:      "data \"hookdeck_gateway_project\" \"test\" {\n  id = \"tm_out\"\n}\n",
				ExpectError: regexp.MustCompile(`Unsupported project type`),
			},
			{
				Config:      byName("deliveries"),
				ExpectError: regexp.MustCompile(`Project not found`),
			},
			{
				Config: byName("alpha"),
				Check:  resource.TestCheckResourceAttr("data.hookdeck_gateway_project.test", "id", "tm_a"),
			},
		},
	})
}

// Renaming a project leaves the settings made elsewhere as they are.
func TestAccMock_GatewayProject_RenameKeepsSettings(t *testing.T) {
	skipUnlessAcc(t)
	m := mockOrganization(t)
	useAPIKey(t, mockOrgKey)
	var projectID string
	prefix := "x-set-elsewhere"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: projectConfig("gamma"),
				Check:  captureID(projectAddr, &projectID),
			},
			{
				PreConfig: func() {
					m.update(func() {
						m.projects[projectID].HeadersPrefix = &prefix
						m.projects[projectID].NotificationMethods = []string{"email"}
					})
				},
				Config: projectConfig("gamma-renamed"),
				Check: func(_ *terraform.State) error {
					var err error
					m.update(func() {
						p := m.projects[projectID]
						if p.Name != "gamma-renamed" || p.HeadersPrefix == nil || *p.HeadersPrefix != prefix || len(p.NotificationMethods) != 1 {
							err = fmt.Errorf("project after rename: %+v", p)
						}
					})
					return err
				},
			},
		},
	})
}
