package provider_test

import (
	"fmt"
	"net/http"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const outpostDataBlock = `
data "hookdeck_outpost_project" "test" {
  name = hookdeck_outpost_project.test.name
}
`

func outpostLifecycleConfig(name, retryInterval, topics string) string {
	if topics != "" {
		topics = fmt.Sprintf("    TOPICS                      = %q\n", topics)
	}
	return fmt.Sprintf(`resource "hookdeck_outpost_project" "test" {
  name = %q
  config = {
%s    RETRY_INTERVAL_SECONDS      = %q
    OTEL_SERVICE_NAME           = "tf-acc"
    OTEL_EXPORTER_OTLP_ENDPOINT = "https://otel.invalid"
  }
  sensitive_config = {
    OTEL_EXPORTER_OTLP_HEADERS = "x-acc=1"
  }
}
`, name, topics, retryInterval) + outpostDataBlock
}

// One real Outpost project for the whole test: create with the data
// source, update, import, drift, destroy.
func TestAccOutpostProject_Lifecycle(t *testing.T) {
	orgKey := acc.OrgKey(t)
	name := "tf-acc-outpost-" + acc.RandSuffix()
	var projectID string

	checkConfig := func(key, want string) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			status, body := acc.OutpostRequest(t, projectID, http.MethodGet, "/config", nil)
			if status != http.StatusOK || body[key] != want {
				return fmt.Errorf("GET /config: status %d, %s = %v, want %s", status, key, body[key], want)
			}
			return nil
		}
	}
	patch := func(values map[string]any) func() {
		return func() {
			if status, body := acc.OutpostRequest(t, projectID, http.MethodPatch, "/config", values); status != http.StatusOK {
				t.Fatalf("PATCH /config: status %d: %v", status, body)
			}
		}
	}
	original := outpostLifecycleConfig(name, "45", "orders.created,orders.updated")
	updated := outpostLifecycleConfig(name+"-renamed", "60", "")

	acc.TestAs(t, acc.OrgNoProject(t), resource.TestCase{
		CheckDestroy: func(_ *terraform.State) error {
			if projectID == "" {
				return nil
			}
			status, _ := acc.Request(t, orgKey, "", http.MethodGet, "/projects/"+projectID, nil)
			if status != http.StatusNotFound && status != http.StatusGone {
				return fmt.Errorf("project %s still exists (status %d)", projectID, status)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: original,
				Check: resource.ComposeAggregateTestCheckFunc(
					captureID(outpostAddr, &projectID),
					resource.TestCheckResourceAttr(outpostAddr, "type", "outpost"),
					resource.TestCheckResourceAttr(outpostAddr, "config.%", "4"),
					resource.TestCheckResourceAttr(outpostAddr, "config.TOPICS", "orders.created,orders.updated"),
					resource.TestCheckResourceAttr(outpostAddr, "config.RETRY_INTERVAL_SECONDS", "45"),
					resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.%", "1"),
					resource.TestCheckResourceAttr(outpostAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "x-acc=1"),
					resource.TestCheckResourceAttrPair(outpostDataAddr, "id", outpostAddr, "id"),
					resource.TestCheckResourceAttr(outpostDataAddr, "type", "outpost"),
					resource.TestCheckResourceAttr(outpostDataAddr, "config.MAX_RETRY_LIMIT", "10"),
					resource.TestCheckResourceAttr(outpostDataAddr, "config.TOPICS", "orders.created,orders.updated"),
					resource.TestCheckNoResourceAttr(outpostDataAddr, "config.OTEL_EXPORTER_OTLP_HEADERS"),
					resource.TestCheckResourceAttr(outpostDataAddr, "sensitive_config.OTEL_EXPORTER_OTLP_HEADERS", "x-acc=1"),
				),
			},
			{
				Config: updated,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(outpostAddr, plancheck.ResourceActionUpdate),
				}},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkIDEquals(outpostAddr, &projectID),
					resource.TestCheckResourceAttr(outpostAddr, "name", name+"-renamed"),
					resource.TestCheckResourceAttr(outpostAddr, "config.%", "3"),
					resource.TestCheckNoResourceAttr(outpostAddr, "config.TOPICS"),
					resource.TestCheckResourceAttr(outpostAddr, "config.RETRY_INTERVAL_SECONDS", "60"),
					resource.TestCheckResourceAttr(outpostDataAddr, "config.RETRY_INTERVAL_SECONDS", "60"),
					checkConfig("RETRY_INTERVAL_SECONDS", "60"),
				),
			},
			{Config: updated, PlanOnly: true},
			{
				ResourceName:            outpostAddr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config", "sensitive_config", "timeouts"},
			},
			{
				PreConfig: patch(map[string]any{"MAX_RETRY_LIMIT": "5"}),
				Config:    updated,
				PlanOnly:  true,
			},
			{
				PreConfig:          patch(map[string]any{"RETRY_INTERVAL_SECONDS": "120"}),
				Config:             updated,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(outpostAddr, "config.RETRY_INTERVAL_SECONDS", "60"),
					checkConfig("RETRY_INTERVAL_SECONDS", "60"),
				),
			},
		},
	})
}
