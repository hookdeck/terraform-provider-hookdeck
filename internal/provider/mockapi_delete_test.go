package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const allResourcesConfig = `
resource "hookdeck_gateway_source" "test" {
  name = "mock-src"
  type = "HTTP"
}

resource "hookdeck_gateway_source_auth" "test" {
  source_id = hookdeck_gateway_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "secret"
  })
}

resource "hookdeck_gateway_destination" "test" {
  name   = "mock-dst"
  type   = "HTTP"
  config = jsonencode({ url = "https://example.com/hook" })
}

resource "hookdeck_gateway_transformation" "test" {
  name = "mock-trn"
  code = "addHandler(\"transform\", (request, context) => request);"
}

resource "hookdeck_gateway_connection" "test" {
  name           = "mock-con"
  source_id      = hookdeck_gateway_source.test.id
  destination_id = hookdeck_gateway_destination.test.id
}
`

const legacySourceConfig = `
resource "hookdeck_source" "test" {
  name = "mock-src"
  type = "HTTP"
}
`

// faultResource is the resource whose delete sends each fault's request.
var faultResource = map[string]string{
	"DELETE sources":         "hookdeck_gateway_source.test",
	"DELETE destinations":    "hookdeck_gateway_destination.test",
	"DELETE connections":     "hookdeck_gateway_connection.test",
	"DELETE transformations": "hookdeck_gateway_transformation.test",
	"PUT sources":            "hookdeck_gateway_source_auth.test",
}

// destroyWithFault applies config, then destroys it without refreshing while
// the API answers the faulty request with status, so the destroy reaches the
// resource's own delete. It expects the destroy to succeed unless wantError
// is set, in which case the resource must still be in state afterwards and
// the fault is lifted so the harness can clean up.
func destroyWithFault(t *testing.T, m *mockAPI, config, fault string, status int, wantError *regexp.Regexp) {
	t.Helper()
	acc.UseAPIKey(t, mockProjectKeyA)

	faultHits := func() int {
		m.mu.Lock()
		defer m.mu.Unlock()
		return m.faultHits[fault]
	}
	steps := []resource.TestStep{
		{Config: config},
		{
			PreConfig:   func() { m.update(func() { m.faults[fault] = status }) },
			Config:      config,
			Destroy:     true,
			ExpectError: wantError,
		},
	}
	if wantError != nil {
		steps = append(steps, resource.TestStep{
			PreConfig: func() {
				if faultHits() == 0 {
					t.Errorf("the destroy never sent %q", fault)
				}
				m.update(func() { delete(m.faults, fault) })
			},
			Config: config,
			// The resources the destroy did delete come back; the one it
			// failed on must still be in state.
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(faultResource[fault], plancheck.ResourceActionNoop)},
			},
		})
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: steps,
	})
	if wantError == nil && faultHits() == 0 {
		t.Errorf("the destroy never sent %q", fault)
	}
}

// Deleting a resource the API says is gone succeeds, for 404 and 410 alike.
// The destroy skips the refresh, which is what a deletion between refresh
// and delete looks like.
func TestAccMock_DeleteAlreadyGone(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, collection := range []string{"sources", "destinations", "connections", "transformations"} {
		for _, status := range []int{404, 410} {
			t.Run(fmt.Sprintf("%s/%d", collection, status), func(t *testing.T) {
				m := mockOrganization(t)
				destroyWithFault(t, m, allResourcesConfig, "DELETE "+collection, status, nil)
			})
		}
	}
}

func TestAccMock_DeleteServerErrorFails(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for collection, singular := range map[string]string{"sources": "source", "destinations": "destination", "connections": "connection", "transformations": "transformation"} {
		t.Run(collection, func(t *testing.T) {
			m := mockOrganization(t)
			destroyWithFault(t, m, allResourcesConfig, "DELETE "+collection, 500, regexp.MustCompile("Error deleting "+singular))
		})
	}
}

// source_auth deletes by reading its source and writing it back without
// auth. The source being gone at the read or at the write means there is
// nothing left to remove.
func TestAccMock_SourceAuthDeleteWriteAlreadyGone(t *testing.T) {
	acc.SkipUnlessAcc(t)
	for _, status := range []int{404, 410} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			m := mockOrganization(t)
			destroyWithFault(t, m, allResourcesConfig, "PUT sources", status, nil)
		})
	}
}

func TestAccMock_SourceAuthDeleteReadAlreadyGone(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	destroyWithFault(t, m, allResourcesConfig, "GET sources", 404, nil)
}

func TestAccMock_SourceAuthDeleteWriteServerErrorFails(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	destroyWithFault(t, m, allResourcesConfig, "PUT sources", 500, regexp.MustCompile("Error deleting source auth"))
}

// The v2 names share the delete code of the new ones.
func TestAccMock_LegacySourceDeleteAlreadyGone(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	destroyWithFault(t, m, legacySourceConfig, "DELETE sources", 410, nil)
}

// A 404 that refuses the project is not a resource that is gone: deleting
// with an organization key whose project no longer exists fails and keeps
// the resource in state.
func TestAccMock_DeleteProjectRefusedFails(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockOrgKey)
	var removed *mockProject
	config := sourceConfig(t, "mock", "tm_a")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		// A refresh would hit the same refusal in Read; skipping it makes
		// the destroy send the DELETE.
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: func() {
					m.update(func() {
						removed = m.projects["tm_a"]
						delete(m.projects, "tm_a")
					})
				},
				Config:      config,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s)cannot access project tm_a.*does not exist`),
			},
			{
				PreConfig: func() {
					deletes := 0
					for _, r := range m.requestsTo("/sources/") {
						if r.method == http.MethodDelete {
							deletes++
						}
					}
					if deletes == 0 {
						t.Error("the destroy never sent a DELETE")
					}
					m.update(func() { m.projects["tm_a"] = removed })
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}
