package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// A key the provider does not know is still sent; the API decides.
func TestAccMock_OutpostProject_UnknownKeyIsSentAsIs(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config: outpostConfig("deliveries", "  config = {\n    EXTRA_FUTURE_KEY = \"x\"\n  }\n"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr(outpostAddr, "config.EXTRA_FUTURE_KEY", "x"),
				checkSentPatches(m, map[string]*string{"EXTRA_FUTURE_KEY": ptr("x")}),
			),
		}},
	})
}

func TestAccMock_OutpostProject_PlanErrors(t *testing.T) {
	acc.SkipUnlessAcc(t)
	mockOutpostOrganization(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      outpostConfig("deliveries", "  config = {\n    OTEL_EXPORTER_OTLP_HEADERS = \"k=v\"\n  }\n"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)Credential key in config.*sensitive_config`),
			},
			{
				Config:      outpostConfig("deliveries", "  sensitive_config = {\n    OPERATOR_EVENTS_TOPICS = \"x\"\n  }\n"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`OPERATOR_EVENTS_TOPICS cannot be set here`),
			},
			{
				Config:      outpostConfig("deliveries", "  config = {\n    TOPICS = \"a\"\n  }\n  sensitive_config = {\n    TOPICS = \"a\"\n  }\n"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`TOPICS is set in both config and sensitive_config`),
			},
		},
	})
}

// Importing a project of another type is refused, and a data source named
// like a gateway project finds the Outpost project.
func TestAccMock_OutpostProject_OtherProjectTypes(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProject("tm_gw", "shared", "event_gateway")
	m.addProject("tm_ob", "shared", "outpost")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        outpostConfig("shared", ""),
				ResourceName:  outpostAddr,
				ImportState:   true,
				ImportStateId: "tm_gw",
				ExpectError:   regexp.MustCompile(`(?s)Unsupported project type.*event_gateway`),
			},
			{
				Config:      "data \"hookdeck_outpost_project\" \"test\" {\n  id = \"tm_gw\"\n}\n",
				ExpectError: regexp.MustCompile(`Unsupported project type`),
			},
			{
				Config: "data \"hookdeck_outpost_project\" \"test\" {\n  name = \"shared\"\n}\n",
				Check:  resource.TestCheckResourceAttr(outpostDataAddr, "id", "tm_ob"),
			},
		},
	})
}

func TestAccMock_OutpostProject_ProjectKeyIsRefused(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOrganization(t)
	acc.UseAPIKey(t, mockProjectKeyA)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      outpostConfig("deliveries", outpostBasic),
			ExpectError: regexp.MustCompile(`Organization API key required`),
		}},
	})
	if got := m.requestsTo("/config"); len(got) != 0 {
		t.Errorf("requests to /config: %+v", got)
	}
	if got := countRequests(m, http.MethodPost, "/projects"); got != 0 {
		t.Errorf("%d POST /projects requests, want none", got)
	}
}

// The data source returns the whole configuration, Hookdeck defaults
// included. Credential keys and keys the provider does not know are only in
// sensitive_config.
func TestAccMock_OutpostProject_DataSource(t *testing.T) {
	acc.SkipUnlessAcc(t)
	m := mockOutpostOrganization(t)
	m.addProject("tm_out", "deliveries", "outpost")
	m.update(func() {
		m.outposts["tm_out"].overrides = map[string]string{
			"TOPICS":                     "a,b",
			"OTEL_EXPORTER_OTLP_HEADERS": "k=v",
			"DESTINATIONS_PROXY_URL":     "http://user:pass@proxy:8080",
			"EXTRA_FUTURE_KEY":           "future",
		}
	})
	m.addKey(mockOutpostKeyName, mockKey{org: mockOrg, project: "tm_out"})

	checks := resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(outpostDataAddr, "id", "tm_out"),
		resource.TestCheckResourceAttr(outpostDataAddr, "type", "outpost"),
		resource.TestCheckResourceAttr(outpostDataAddr, "config.TOPICS", "a,b"),
		resource.TestCheckResourceAttr(outpostDataAddr, "config.MAX_RETRY_LIMIT", "10"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "config.PORTAL_LOGO"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "sensitive_config.PORTAL_LOGO"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "config.OTEL_EXPORTER_OTLP_HEADERS"),
		resource.TestCheckResourceAttr(outpostDataAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "k=v"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "config.DESTINATIONS_PROXY_URL"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "config.DESTINATIONS_WEBHOOK_PROXY_URL"),
		resource.TestCheckResourceAttr(outpostDataAddr, "sensitive_config.DESTINATIONS_PROXY_URL", "http://user:pass@proxy:8080"),
		resource.TestCheckResourceAttr(outpostDataAddr, "sensitive_config.DESTINATIONS_WEBHOOK_PROXY_URL", "http://user:pass@proxy:8080"),
		resource.TestCheckNoResourceAttr(outpostDataAddr, "config.EXTRA_FUTURE_KEY"),
		resource.TestCheckResourceAttr(outpostDataAddr, "sensitive_config.EXTRA_FUTURE_KEY", "future"),
	)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acc.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: "data \"hookdeck_outpost_project\" \"test\" {\n  id = \"tm_out\"\n}\n",
				Check:  checks,
			},
			{
				Config: "data \"hookdeck_outpost_project\" \"test\" {\n  name = \"deliveries\"\n}\n",
				Check:  checks,
			},
			{
				PreConfig: func() { acc.UseAPIKey(t, mockOutpostKeyName) },
				Config:    "data \"hookdeck_outpost_project\" \"test\" {\n  id = \"tm_out\"\n}\n",
				Check:     checks,
			},
		},
	})
}
