package provider_test

import (
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The data source returns the domain and its status, and null for a project
// without one.
func TestAccMock_OutpostProject_DataSourceDomain(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProject("tm_with", "with-domain", "outpost")
	m.addProject("tm_without", "without-domain", "outpost")
	m.update(func() {
		m.outposts["tm_with"].domain = &mockDomain{hostname: domainA, active: true}
		m.hostnames[domainA] = true
	})
	addr := func(name string) string { return "data.hookdeck_outpost_project." + name }

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `data "hookdeck_outpost_project" "with" {
  name = "with-domain"
}

data "hookdeck_outpost_project" "without" {
  name = "without-domain"
}
`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(addr("with"), "portal_custom_domain", domainA),
				resource.TestCheckResourceAttr(addr("with"), statusAttr+".status", "active"),
				resource.TestCheckResourceAttr(addr("with"), statusAttr+".ssl_status", "active"),
				resource.TestCheckResourceAttr(addr("with"), statusAttr+".verification_records.#", "2"),
				resource.TestCheckResourceAttr(addr("with"), statusAttr+".verification_records.0.name", "_acme-challenge."+domainA),
				resource.TestCheckNoResourceAttr(addr("without"), "portal_custom_domain"),
				resource.TestCheckNoResourceAttr(addr("without"), statusAttr+".status"),
			),
		}},
	})
}
