package provider_test

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Create waits until the project is provisioned, then sends one PATCH with
// both maps. State holds the keys the user set and nothing else.
func TestAccMock_OutpostProject_CreateWaitsAndSendsOnePatch(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.update(func() { m.provisionReads = 2 })

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: outpostConfig("deliveries", outpostBasic),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(outpostAddr, "type", "outpost"),
				resource.TestCheckResourceAttrSet(outpostAddr, "id"),
				resource.TestCheckResourceAttr(outpostAddr, "config.%", "2"),
				resource.TestCheckResourceAttr(outpostAddr, "config.TOPICS", "a,b"),
				resource.TestCheckResourceAttr(outpostAddr, "config.RETRY_INTERVAL_SECONDS", "60"),
				resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.%", "1"),
				resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "k=v"),
				checkSentPatches(m, map[string]*string{
					"TOPICS": ptr("a,b"), "RETRY_INTERVAL_SECONDS": ptr("60"), "OTEL_EXPORTER_OTLP_HEADERS": ptr("k=v"),
				}),
				func(_ *terraform.State) error {
					var reads int
					for _, r := range m.requestsTo("/config") {
						if r.method == http.MethodGet {
							reads++
						}
					}
					if reads < 3 {
						return fmt.Errorf("%d GET /config before the project was ready, want at least 3", reads)
					}
					return nil
				},
			),
		}},
	})
}

func TestAccMock_OutpostProject_CreateWithoutConfigSendsNoPatch(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: outpostConfig("deliveries", ""),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet(outpostAddr, "id"),
				resource.TestCheckNoResourceAttr(outpostAddr, "config.%"),
				checkSentPatches(m),
			),
		}},
	})
}

// A changed key and a removed key are one PATCH. A rename sends no PATCH,
// and an apply with nothing to do sends no write at all.
func TestAccMock_OutpostProject_Update(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var patches, puts int
	remember := func() {
		patches = countRequests(m, http.MethodPatch, "/config")
		puts = countRequests(m, http.MethodPut, "/projects")
	}
	body := func(retry string, extra string) string {
		return fmt.Sprintf("  config = {\n    RETRY_INTERVAL_SECONDS = %q\n%s  }\n", retry, extra)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: outpostConfig("deliveries", body("60", "    TOPICS = \"a\"\n"))},
			{
				PreConfig: remember,
				Config:    outpostConfig("deliveries", body("90", "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(outpostAddr, "config.%", "1"),
					resource.TestCheckResourceAttr(outpostAddr, "config.RETRY_INTERVAL_SECONDS", "90"),
					func(_ *terraform.State) error {
						patches := m.sentPatches()
						want := map[string]*string{"RETRY_INTERVAL_SECONDS": ptr("90"), "TOPICS": nil}
						if len(patches) != 2 || !reflect.DeepEqual(patches[1], want) {
							return fmt.Errorf("PATCH bodies = %s", formatPatches(patches))
						}
						return nil
					},
				),
			},
			{
				PreConfig: remember,
				Config:    outpostConfig("deliveries-renamed", body("90", "")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(outpostAddr, "name", "deliveries-renamed"),
					func(_ *terraform.State) error {
						if got := countRequests(m, http.MethodPatch, "/config") - patches; got != 0 {
							return fmt.Errorf("%d PATCH /config requests for a rename, want none", got)
						}
						if got := countRequests(m, http.MethodPut, "/projects") - puts; got != 1 {
							return fmt.Errorf("%d PUT /projects requests for a rename, want 1", got)
						}
						return nil
					},
				),
			},
			{
				PreConfig: remember,
				Config:    outpostConfig("deliveries-renamed", body("90", "")),
				Check:     checkNoNewWrites(m, &patches, &puts),
			},
		},
	})
}

// A managed key edited elsewhere shows in the plan; a key the user does not
// manage never does.
func TestAccMock_OutpostProject_Drift(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	config := outpostConfig("deliveries", "  config = {\n    TOPICS = \"a\"\n  }\n")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(outpostAddr, &projectID)},
			{
				PreConfig:          func() { m.update(func() { m.outposts[projectID].overrides["TOPICS"] = "x" }) },
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					m.update(func() {
						m.outposts[projectID].overrides["TOPICS"] = "a"
						m.outposts[projectID].overrides["OTEL_SERVICE_NAME"] = "edited"
						m.outposts[projectID].overrides["MAX_RETRY_LIMIT"] = "99"
					})
				},
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// Reading a project that is not provisioned keeps the configuration in
// state: no error, no diff.
func TestAccMock_OutpostProject_ReadNotProvisionedKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	config := outpostConfig("deliveries", outpostBasic)
	setProvisioned := func(reads int) func() {
		return func() { m.update(func() { m.outposts[projectID].notProvisioned = reads }) }
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(outpostAddr, &projectID)},
			{PreConfig: setProvisioned(-1), Config: config, PlanOnly: true},
			{PreConfig: setProvisioned(0), Config: config, PlanOnly: true},
		},
	})
}

// Moving a key between config and sensitive_config with the same value
// sends nothing and leaves an empty plan.
func TestAccMock_OutpostProject_KeyMovedBetweenMaps(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var patches int
	moved := outpostConfig("deliveries", "  sensitive_config = {\n    TOPICS = \"a\"\n  }\n")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: outpostConfig("deliveries", "  config = {\n    TOPICS = \"a\"\n  }\n")},
			{
				PreConfig: func() { patches = countRequests(m, http.MethodPatch, "/config") },
				Config:    moved,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.TOPICS", "a"),
					resource.TestCheckNoResourceAttr(outpostAddr, "config.TOPICS"),
					func(_ *terraform.State) error {
						if got := countRequests(m, http.MethodPatch, "/config") - patches; got != 0 {
							return fmt.Errorf("%d PATCH /config requests for a moved key, want none", got)
						}
						return nil
					},
				),
			},
			{Config: moved, PlanOnly: true},
		},
	})
}

// Importing by ID fills the project and leaves the config maps null. The
// first apply then sends nothing when the configuration already matches.
func TestAccMock_OutpostProject_Import(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)

	t.Run("verify", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acc.ProviderFactories,
			Steps: []resource.TestStep{
				{Config: outpostConfig("deliveries", outpostBasic)},
				{
					Config:                  outpostConfig("deliveries", outpostBasic),
					ResourceName:            outpostAddr,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"config", "sensitive_config", "timeouts"},
				},
			},
		})
	})

	t.Run("first apply sends nothing", func(t *testing.T) {
		m.addProject("tm_imp", "imported", "outpost")
		m.update(func() {
			m.outposts["tm_imp"].overrides = map[string]string{"TOPICS": "a,b", "OTEL_EXPORTER_OTLP_HEADERS": "k=v"}
		})
		before := len(m.sentPatches())
		config := `import {
  to = hookdeck_outpost_project.test
  id = "tm_imp"
}
` + outpostConfig("imported", "  config = {\n    TOPICS = \"a,b\"\n  }\n  sensitive_config = {\n    OTEL_EXPORTER_OTLP_HEADERS = \"k=v\"\n  }\n")

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acc.ProviderFactories,
			Steps: []resource.TestStep{{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(outpostAddr, "id", "tm_imp"),
					resource.TestCheckResourceAttr(outpostAddr, "config.TOPICS", "a,b"),
					resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "k=v"),
					func(_ *terraform.State) error {
						if got := len(m.sentPatches()) - before; got != 0 {
							return fmt.Errorf("%d PATCH requests, want none", got)
						}
						return nil
					},
				),
			}},
		})
	})
}

// A project that never finishes provisioning fails the create, keeps its ID
// in state, and is replaced by the next apply.
func TestAccMock_OutpostProject_ProvisioningTimeoutIsReplaced(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.update(func() { m.provisionReads = -1 })
	var firstID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      outpostConfig("deliveries", "  timeouts {\n    create = \"5s\"\n  }\n"),
				ExpectError: regexp.MustCompile(`(?s)Outpost project tm_new\d+ was created but is not ready after 5s`),
			},
			{
				PreConfig: func() {
					m.update(func() {
						m.provisionReads = 0
						for id, p := range m.projects {
							if p.Type == "outpost" {
								firstID = id
							}
						}
					})
				},
				Config:           outpostConfig("deliveries", ""),
				ConfigPlanChecks: replacePlan(),
				Check: func(s *terraform.State) error {
					if got := s.RootModule().Resources[outpostAddr].Primary.ID; got == firstID {
						return fmt.Errorf("the tainted project %s was not replaced", got)
					}
					return nil
				},
			},
		},
	})
	if m.projects[firstID] != nil {
		t.Errorf("project %s was not deleted by the replacement", firstID)
	}
}

// A config the API rejects at create leaves the project in state, tainted.
func TestAccMock_OutpostProject_ConfigRejectedAtCreateIsReplaced(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var firstID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      outpostConfig("deliveries", "  config = {\n    NOT_A_KEY = \"x\"\n  }\n"),
				ExpectError: regexp.MustCompile(`(?s)NOT_A_KEY.*is not allowed`),
			},
			{
				PreConfig: func() {
					m.update(func() {
						for id, p := range m.projects {
							if p.Type == "outpost" {
								firstID = id
							}
						}
					})
				},
				Config:           outpostConfig("deliveries", "  config = {\n    TOPICS = \"a\"\n  }\n"),
				ConfigPlanChecks: replacePlan(),
				Check: func(s *terraform.State) error {
					if got := s.RootModule().Resources[outpostAddr].Primary.ID; got == firstID {
						return fmt.Errorf("the tainted project %s was not replaced", got)
					}
					return nil
				},
			},
		},
	})
}

// A project deleted elsewhere drops out of state and is created again.
func TestAccMock_OutpostProject_DeletedOutsideTerraform(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	config := outpostConfig("deliveries", "  config = {\n    TOPICS = \"a\"\n  }\n")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(outpostAddr, &projectID)},
			{
				PreConfig: func() {
					m.update(func() {
						delete(m.projects, projectID)
						delete(m.outposts, projectID)
					})
				},
				Config:           config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(outpostAddr, plancheck.ResourceActionCreate)}},
				Check:            checkIDChanged(outpostAddr, &projectID),
			},
		},
	})
}

// Destroy deletes the project.
func TestAccMock_OutpostProject_Delete(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: outpostConfig("deliveries", ""),
			Check:  captureID(outpostAddr, &projectID),
		}},
		CheckDestroy: func(_ *terraform.State) error {
			m.mu.Lock()
			defer m.mu.Unlock()
			if _, ok := m.projects[projectID]; ok {
				return fmt.Errorf("project %s still exists", projectID)
			}
			return nil
		},
	})
	deletes := 0
	for _, r := range m.requestsTo("/projects/" + projectID) {
		if r.method == http.MethodDelete {
			deletes++
		}
	}
	if deletes != 1 {
		t.Errorf("%d DELETE /projects/%s requests, want 1", deletes, projectID)
	}
}
