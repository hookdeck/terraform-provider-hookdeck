package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

const (
	domainA    = "portal.example.com"
	domainB    = "deliveries.example.com"
	statusAttr = "portal_custom_domain_status"
)

func outpostDomainConfig(name, hostname string) string {
	if hostname == "" {
		return outpostConfig(name, "")
	}
	return outpostConfig(name, fmt.Sprintf("  portal_custom_domain = %q\n", hostname))
}

func domainCalls(m *mockAPI, method string) int {
	return countRequests(m, method, "/config/custom_domain")
}

func statusIsUnknown() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction(outpostAddr, plancheck.ResourceActionUpdate),
		plancheck.ExpectUnknownValue(outpostAddr, tfjsonpath.New(statusAttr)),
	}}
}

func checkDomain(hostname string, active bool) resource.TestCheckFunc {
	status, ssl := "pending", "pending_validation"
	if active {
		status, ssl = "active", "active"
	}
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(outpostAddr, "portal_custom_domain", hostname),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".status", status),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".cname_target", "portal.outpost.example.net"),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ssl.status", ssl),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ssl.method", "txt"),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ssl.txt_name", "_acme-challenge."+hostname),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ssl.txt_value", "ssl-token"),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ownership_verification.type", "txt"),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ownership_verification.name", "_ownership."+hostname),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".ownership_verification.value", "ownership-token"),
		resource.TestCheckResourceAttr(outpostAddr, statusAttr+".verification_errors.#", "0"),
	)
}

// The domain is added after the configuration. The status has the DNS
// records and follows the domain as it becomes active.
func TestAccMock_OutpostProject_CreateWithDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	config := outpostConfig("deliveries", outpostBasic+"  portal_custom_domain = \""+domainA+"\"\n")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(outpostAddr, &projectID),
					checkDomain(domainA, false),
					func(_ *terraform.State) error {
						if got := domainCalls(m, http.MethodPost); got != 1 {
							return fmt.Errorf("%d POST /config/custom_domain requests, want 1", got)
						}
						if got := len(m.sentPatches()); got != 1 {
							return fmt.Errorf("%d PATCH /config requests, want 1", got)
						}
						return nil
					},
				),
			},
			{
				PreConfig: func() { m.update(func() { m.outposts[projectID].domain.active = true }) },
				Config:    config,
				Check:     checkDomain(domainA, true),
			},
			{Config: config, PlanOnly: true},
		},
	})
}

// A new hostname is DELETE then POST in one update; removing it is a DELETE.
// Renaming the project leaves the status as it is.
func TestAccMock_OutpostProject_ChangeAndRemoveDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var posts, deletes int
	remember := func() { posts, deletes = domainCalls(m, http.MethodPost), domainCalls(m, http.MethodDelete) }
	calls := func(wantPosts, wantDeletes int) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			if p, d := domainCalls(m, http.MethodPost)-posts, domainCalls(m, http.MethodDelete)-deletes; p != wantPosts || d != wantDeletes {
				return fmt.Errorf("%d POST and %d DELETE on the domain, want %d and %d", p, d, wantPosts, wantDeletes)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: outpostDomainConfig("deliveries", domainA), Check: checkDomain(domainA, false)},
			{
				PreConfig:        remember,
				Config:           outpostDomainConfig("deliveries", domainB),
				ConfigPlanChecks: statusIsUnknown(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainB, false),
					calls(1, 1),
					func(_ *terraform.State) error {
						if m.hostnames[domainA] || !m.hostnames[domainB] {
							return fmt.Errorf("registered hostnames = %v", m.hostnames)
						}
						return nil
					},
				),
			},
			{
				PreConfig: remember,
				Config:    outpostDomainConfig("renamed", domainB),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue(outpostAddr, tfjsonpath.New(statusAttr).AtMapKey("status"), knownvalue.StringExact("pending")),
				}},
				Check: calls(0, 0),
			},
			{
				PreConfig: remember,
				Config:    outpostDomainConfig("renamed", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(outpostAddr, "portal_custom_domain"),
					resource.TestCheckNoResourceAttr(outpostAddr, statusAttr+".status"),
					calls(0, 1),
				),
			},
			{Config: outpostDomainConfig("renamed", ""), PlanOnly: true},
		},
	})
}

// A hostname changed or removed outside Terraform shows in the plan, and so
// does a domain added outside Terraform to a project that sets none.
func TestAccMock_OutpostProject_DomainDrift(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	set := func(domain *mockDomain) func() {
		return func() {
			m.update(func() {
				if old := m.outposts[projectID].domain; old != nil {
					delete(m.hostnames, old.hostname)
				}
				m.outposts[projectID].domain = domain
				if domain != nil {
					m.hostnames[domain.hostname] = true
				}
			})
		}
	}
	withDomain := outpostDomainConfig("deliveries", domainA)
	without := outpostDomainConfig("deliveries", "")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: withDomain, Check: captureID(outpostAddr, &projectID)},
			{PreConfig: set(&mockDomain{hostname: domainB}), Config: withDomain, PlanOnly: true, ExpectNonEmptyPlan: true},
			{PreConfig: set(nil), Config: withDomain, PlanOnly: true, ExpectNonEmptyPlan: true},
			{PreConfig: set(nil), Config: withDomain, Check: checkDomain(domainA, false)},
			{PreConfig: set(&mockDomain{hostname: domainB}), Config: without, PlanOnly: true, ExpectNonEmptyPlan: true},
			{
				Config: without,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(outpostAddr, "portal_custom_domain"),
					func(_ *terraform.State) error {
						if m.outposts[projectID].domain != nil {
							return fmt.Errorf("the domain added outside Terraform was not removed")
						}
						return nil
					},
				),
			},
		},
	})
}

// Destroy deletes the domain before the project, so the hostname is free
// again and can be added to another project.
func TestAccMock_OutpostProject_DeleteRemovesDomainFirst(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: outpostDomainConfig("deliveries", domainA),
			Check:  captureID(outpostAddr, &projectID),
		}},
	})

	domainDelete, projectDelete := -1, -1
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.requests {
		switch {
		case r.method == http.MethodDelete && r.path == "/config/custom_domain":
			domainDelete = i
		case r.method == http.MethodDelete && r.path == "/projects/"+projectID:
			projectDelete = i
		}
	}
	if domainDelete < 0 || projectDelete < 0 || domainDelete > projectDelete {
		t.Errorf("DELETE domain at request %d, DELETE project at %d; the domain must go first", domainDelete, projectDelete)
	}
	if m.hostnames[domainA] {
		t.Errorf("%s is still registered after destroy", domainA)
	}
}

// Importing a project reads its domain from the API.
func TestAccMock_OutpostProject_ImportReadsDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProject("tm_dom", "deliveries", "outpost")
	m.update(func() {
		m.outposts["tm_dom"].domain = &mockDomain{hostname: domainA}
		m.hostnames[domainA] = true
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: "import {\n  to = hookdeck_outpost_project.test\n  id = \"tm_dom\"\n}\n" + outpostDomainConfig("deliveries", domainA),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(outpostAddr, tfjsonpath.New("portal_custom_domain"), knownvalue.StringExact(domainA)),
				statecheck.ExpectKnownValue(outpostAddr, tfjsonpath.New(statusAttr).AtMapKey("status"), knownvalue.StringExact("pending")),
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(outpostAddr, plancheck.ResourceActionNoop),
			}},
			Check: func(_ *terraform.State) error {
				if got := domainCalls(m, http.MethodPost) + domainCalls(m, http.MethodDelete); got != 0 {
					return fmt.Errorf("%d domain writes for an imported project, want none", got)
				}
				return nil
			},
		}},
	})
}

// Import by ID, verified against the state of an applied project.
func TestAccMock_OutpostProject_ImportStateVerifyWithDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	mockOutpostOrganization(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: outpostDomainConfig("deliveries", domainA)},
			{
				ResourceName:            outpostAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

// A domain the API refuses on create leaves the project in state, tainted;
// the next apply replaces it.
func TestAccMock_OutpostProject_FailedDomainAddIsReplaced(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.update(func() { m.hostnames[domainA] = true })
	var firstID string
	config := outpostDomainConfig("deliveries", domainA)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`(?s)Error adding portal custom domain.*CUSTOM_DOMAIN_ALREADY_EXISTS`),
			},
			{
				PreConfig: func() {
					m.update(func() {
						delete(m.hostnames, domainA)
						for id, p := range m.projects {
							if p.Type == "outpost" {
								firstID = id
							}
						}
					})
				},
				Config:           config,
				ConfigPlanChecks: replacePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainA, false),
					func(s *terraform.State) error {
						if got := s.RootModule().Resources[outpostAddr].Primary.ID; got == firstID {
							return fmt.Errorf("the tainted project %s was not replaced", got)
						}
						return nil
					},
				),
			},
		},
	})
}

// A hostname change whose add fails leaves the project without a domain; the
// next apply adds the new one without a second delete.
func TestAccMock_OutpostProject_FailedDomainChange(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var posts, deletes int

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: outpostDomainConfig("deliveries", domainA), Check: checkDomain(domainA, false)},
			{
				PreConfig:   func() { m.update(func() { m.hostnames[domainB] = true }) },
				Config:      outpostDomainConfig("deliveries", domainB),
				ExpectError: regexp.MustCompile(`(?s)Error adding portal custom domain.*CUSTOM_DOMAIN_ALREADY_EXISTS`),
			},
			{
				PreConfig: func() {
					m.update(func() { delete(m.hostnames, domainB) })
					posts, deletes = domainCalls(m, http.MethodPost), domainCalls(m, http.MethodDelete)
				},
				Config: outpostDomainConfig("deliveries", domainB),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainB, false),
					func(_ *terraform.State) error {
						if p, d := domainCalls(m, http.MethodPost)-posts, domainCalls(m, http.MethodDelete)-deletes; p != 1 || d != 0 {
							return fmt.Errorf("%d POST and %d DELETE on the domain, want 1 and 0", p, d)
						}
						return nil
					},
				),
			},
		},
	})
}

// The domain route answers for a project that is not provisioned: refresh,
// replacement and destroy work on one whose create timed out before the
// domain was added.
func TestAccMock_OutpostProject_NotProvisionedProjectReplaced(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.update(func() { m.provisionReads = -1 })
	var firstID string
	body := "  portal_custom_domain = \"" + domainA + "\"\n"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      outpostConfig("deliveries", body+"  timeouts {\n    create = \"5s\"\n  }\n"),
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
				Config:           outpostConfig("deliveries", body),
				ConfigPlanChecks: replacePlan(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainA, false),
					func(s *terraform.State) error {
						if got := s.RootModule().Resources[outpostAddr].Primary.ID; got == firstID {
							return fmt.Errorf("the tainted project %s was not replaced", got)
						}
						return nil
					},
				),
			},
		},
	})
}

func setDomainFault(m *mockAPI, projectID *string, method string, status int) func() {
	return func() {
		m.update(func() {
			if status == 0 {
				delete(m.outposts[*projectID].domainFaults, method)
				return
			}
			if m.outposts[*projectID].domainFaults == nil {
				m.outposts[*projectID].domainFaults = map[string]int{}
			}
			m.outposts[*projectID].domainFaults[method] = status
		})
	}
}

// A hostname change whose DELETE of the old domain fails saves what the
// update had already applied (here the rename) and keeps the old domain in
// state. The plans skip the refresh, so they show the state as saved; the next
// apply deletes the old domain and adds the new one.
func TestAccMock_OutpostProject_FailedDomainDeleteOnChangeKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	var posts, deletes int

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		AdditionalCLIOptions:     &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: outpostDomainConfig("deliveries", domainA), Check: captureID(outpostAddr, &projectID)},
			{
				PreConfig:   setDomainFault(m, &projectID, http.MethodDelete, 500),
				Config:      outpostDomainConfig("renamed", domainB),
				ExpectError: regexp.MustCompile(`Error deleting portal custom domain`),
			},
			{
				// State has the new name and the old hostname: no change left for them.
				PreConfig: setDomainFault(m, &projectID, http.MethodDelete, 0),
				Config:    outpostDomainConfig("renamed", domainA),
				PlanOnly:  true,
			},
			{
				PreConfig: func() { posts, deletes = domainCalls(m, http.MethodPost), domainCalls(m, http.MethodDelete) },
				Config:    outpostDomainConfig("renamed", domainB),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainB, false),
					func(_ *terraform.State) error {
						if p, d := domainCalls(m, http.MethodPost)-posts, domainCalls(m, http.MethodDelete)-deletes; p != 1 || d != 1 {
							return fmt.Errorf("%d POST and %d DELETE on the domain, want 1 and 1", p, d)
						}
						return nil
					},
				),
			},
		},
	})
}

// A destroy whose domain DELETE succeeds and whose project DELETE fails keeps
// the project in state; the next refresh finds no domain, so the plan adds it
// back.
func TestAccMock_OutpostProject_FailedProjectDeleteAfterDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	config := outpostDomainConfig("deliveries", domainA)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig:   func() { m.update(func() { m.faults["DELETE projects"] = 500 }) },
				Config:      config,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`Error deleting project`),
			},
			{
				PreConfig: func() {
					m.update(func() {
						delete(m.faults, "DELETE projects")
						if m.hostnames[domainA] {
							t.Errorf("%s is still registered: the domain was not deleted before the project", domainA)
						}
					})
				},
				Config:           config,
				ConfigPlanChecks: statusIsUnknown(),
				Check:            checkDomain(domainA, false),
			},
		},
	})
}

// A refresh whose domain read is refused fails and keeps the state.
func TestAccMock_OutpostProject_DomainReadRefusedKeepsState(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var projectID string
	config := outpostDomainConfig("deliveries", domainA)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: captureID(outpostAddr, &projectID)},
			{
				PreConfig:   setDomainFault(m, &projectID, http.MethodGet, 403),
				Config:      config,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Error reading portal custom domain`),
			},
			{
				PreConfig: setDomainFault(m, &projectID, http.MethodGet, 0),
				Config:    config,
				PlanOnly:  true,
			},
		},
	})
}

// The API's refusal of a hostname is reported as it comes, with no check or
// rewrite by the provider.
func TestAccMock_OutpostProject_HostnameRefusedByAPI(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.update(func() { m.refuseHostnames = true })

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      outpostDomainConfig("deliveries", "Not A Host"),
			ExpectError: regexp.MustCompile(`(?s)Error adding portal custom domain.*must\s+be\s+a\s+valid\s+hostname`),
		}},
	})
	if got := domainCalls(m, http.MethodPost); got != 1 {
		t.Errorf("%d POST /config/custom_domain requests, want 1", got)
	}
}

// A hostname that is unknown at plan time and resolves to the current one
// changes nothing: the status is read, and the domain is not touched.
func TestAccMock_OutpostProject_UnknownHostnameUnchanged(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	var posts, deletes int
	// Changing the trigger replaces the hostname resource, so its output is
	// unknown at plan time and the same value after apply.
	config := func(trigger string) string {
		return fmt.Sprintf(`resource "terraform_data" "trigger" {
  input = %q
}

resource "terraform_data" "hostname" {
  input = %q

  lifecycle {
    replace_triggered_by = [terraform_data.trigger]
  }
}

resource "hookdeck_outpost_project" "test" {
  name                 = "deliveries"
  portal_custom_domain = terraform_data.hostname.output
}
`, trigger, domainA)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("one"), Check: checkDomain(domainA, false)},
			{
				PreConfig: func() { posts, deletes = domainCalls(m, http.MethodPost), domainCalls(m, http.MethodDelete) },
				Config:    config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectUnknownValue(outpostAddr, tfjsonpath.New("portal_custom_domain")),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkDomain(domainA, false),
					func(_ *terraform.State) error {
						if p, d := domainCalls(m, http.MethodPost)-posts, domainCalls(m, http.MethodDelete)-deletes; p != 0 || d != 0 {
							return fmt.Errorf("%d POST and %d DELETE on the domain, want none", p, d)
						}
						return nil
					},
				),
			},
		},
	})
}
