package connection_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"terraform-provider-hookdeck/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"hookdeck": providerserver.NewProtocol6WithError(provider.New("test")()),
	}

	// Common test JSON payload for filter rules.
	testFilterStatusJSON = `{"data":{"attributes":{"payload":{"data":{"attributes":{"status":{"$or":["completed","failed","approved","declined","needs_review"]}}}}}}}`
)

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HOOKDECK_API_KEY"); v == "" {
		t.Fatal("HOOKDECK_API_KEY must be set for acceptance tests")
	}
}

func loadTestConfigFormatted(filename string, args ...interface{}) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	// Format the template with provided arguments
	return fmt.Sprintf(string(content), args...)
}

func loadTestConfigFormattedWithUpdate(filename string, rName string, oldName string, newName string) string {
	config := loadTestConfigFormatted(filename, rName)
	// Replace the connection name for update tests
	return strings.ReplaceAll(config, oldName, newName)
}

func TestAccConnectionResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: loadTestConfigFormatted("basic.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "source_id"),
					resource.TestCheckResourceAttrSet(resourceName, "destination_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "team_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: loadTestConfigFormattedWithUpdate("basic.tf", rName, fmt.Sprintf("test-connection-%s", rName), fmt.Sprintf("test-connection-updated-%s", rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-updated-%s", rName)),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccConnectionResourceWithRules(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with filter rule
			{
				Config: loadTestConfigFormatted("with_filter_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-filter-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.headers.json", `{"x-api-key":"secret"}`),
				),
			},
			// Update with retry rule
			{
				Config: loadTestConfigFormatted("with_retry_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-retry-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.retry_rule.strategy", "exponential"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.retry_rule.count", "5"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.retry_rule.interval", "1000"),
				),
			},
			// Update with delay rule
			{
				Config: loadTestConfigFormatted("with_delay_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-delay-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.delay_rule.delay", "5000"),
				),
			},
			// Update with transform rule
			{
				Config: loadTestConfigFormatted("with_transform_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-transform-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "rules.0.transform_rule.transformation_id"),
				),
			},
			// Update with deduplicate rule
			{
				Config: loadTestConfigFormatted("with_deduplicate_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-deduplicate-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.window", "60000"),
				),
			},
			// Update with multiple rules
			{
				Config: loadTestConfigFormatted("with_multiple_rules.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-multi-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "3"),
				),
			},
		},
	})
}

func TestAccConnectionResourceDeduplicationRule(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test exact deduplication (no fields)
			{
				Config: loadTestConfigFormatted("with_deduplicate_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-deduplicate-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.window", "60000"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields"),
				),
			},
			// Test field-based deduplication with include_fields
			{
				Config: loadTestConfigFormatted("with_deduplicate_include_fields.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-deduplicate-include-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.window", "30000"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields.0", "body.id"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields.1", "headers.x-request-id"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields"),
				),
			},
			// Test field-based deduplication with exclude_fields
			{
				Config: loadTestConfigFormatted("with_deduplicate_exclude_fields.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-deduplicate-exclude-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.window", "45000"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields.0", "body.timestamp"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields.1", "headers.x-trace-id"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields"),
				),
			},
			// Test switching back to exact deduplication
			{
				Config: loadTestConfigFormatted("with_deduplicate_rule.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-deduplicate-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.deduplicate_rule.window", "60000"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.include_fields"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule.exclude_fields"),
				),
			},
		},
	})
}

func TestAccConnectionResourceDeduplicationValidation(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test window validation - too small (< 1000ms)
			{
				Config:      loadTestConfigFormatted("with_deduplicate_invalid_window.tf", rName, 500),
				ExpectError: regexp.MustCompile(`value must be between 1000 and\s+3600000, got: 500`),
			},
			// Test window validation - too large (> 3600000ms)
			{
				Config:      loadTestConfigFormatted("with_deduplicate_invalid_window.tf", rName, 3600001),
				ExpectError: regexp.MustCompile(`value must be between 1000 and\s+3600000, got: 3600001`),
			},
			// Test mutual exclusivity - both include_fields and exclude_fields set
			{
				Config:      loadTestConfigFormatted("with_deduplicate_both_fields.tf", rName),
				ExpectError: regexp.MustCompile(`Only one of \[include_fields, exclude_fields\] can be specified`),
			},
		},
	})
}

func TestAccConnectionResourceRuleOrdering(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test all rule types maintain order: filter -> deduplicate -> delay -> transform -> retry
			{
				Config: loadTestConfigFormatted("with_all_rules_ordered.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-ordered-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "5"),
					// Verify exact order is preserved
					// Position 0: Filter rule
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.headers.json", `{"x-webhook-type":"order.created"}`),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.deduplicate_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.delay_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.transform_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.retry_rule"),
					// Position 1: Deduplicate rule
					resource.TestCheckResourceAttr(resourceName, "rules.1.deduplicate_rule.window", "30000"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.deduplicate_rule.include_fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.deduplicate_rule.include_fields.0", "body.order_id"),
					resource.TestCheckResourceAttr(resourceName, "rules.1.deduplicate_rule.include_fields.1", "body.customer_id"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.filter_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.delay_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.transform_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.retry_rule"),
					// Position 2: Delay rule
					resource.TestCheckResourceAttr(resourceName, "rules.2.delay_rule.delay", "2000"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.2.filter_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.2.deduplicate_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.2.transform_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.2.retry_rule"),
					// Position 3: Transform rule
					resource.TestCheckResourceAttrSet(resourceName, "rules.3.transform_rule.transformation_id"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.3.filter_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.3.deduplicate_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.3.delay_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.3.retry_rule"),
					// Position 4: Retry rule
					resource.TestCheckResourceAttr(resourceName, "rules.4.retry_rule.strategy", "exponential"),
					resource.TestCheckResourceAttr(resourceName, "rules.4.retry_rule.count", "3"),
					resource.TestCheckResourceAttr(resourceName, "rules.4.retry_rule.interval", "5000"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.4.filter_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.4.deduplicate_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.4.delay_rule"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.4.transform_rule"),
				),
			},
			// Update with different order: retry -> transform -> delay -> filter
			{
				Config: loadTestConfigFormatted("with_rules_reordered.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-reordered-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "4"),
					// Verify new order is applied
					// Position 0: Now Retry rule
					resource.TestCheckResourceAttr(resourceName, "rules.0.retry_rule.strategy", "exponential"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.0.filter_rule"),
					// Position 1: Now Transform rule
					resource.TestCheckResourceAttrSet(resourceName, "rules.1.transform_rule.transformation_id"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.1.retry_rule"),
					// Position 2: Now Delay rule
					resource.TestCheckResourceAttr(resourceName, "rules.2.delay_rule.delay", "2000"),
					resource.TestCheckNoResourceAttr(resourceName, "rules.2.transform_rule"),
					// Position 3: Now Filter rule
					resource.TestCheckResourceAttr(resourceName, "rules.3.filter_rule.headers.json", `{"x-webhook-type":"order.created"}`),
					resource.TestCheckNoResourceAttr(resourceName, "rules.3.retry_rule"),
				),
			},
		},
	})
}

// TestAccConnectionResourceFilterJSONFormattingRawWorkaround tests JSON formatting with jsonencode(jsondecode(...)).
// This workaround should work correctly as jsonencode normalizes the JSON.
func TestAccConnectionResourceFilterJSONFormattingRawWorkaround(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted("with_json_formatting_decode.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-json-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.body.json", testFilterStatusJSON),
				),
			},
			// Re-apply to ensure no drift
			{
				Config:   loadTestConfigFormatted("with_json_formatting_decode.tf", rName),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.body.json", testFilterStatusJSON),
				),
			},
		},
	})
}

// TestAccConnectionResourceFilterJSONFormattingRaw tests raw JSON in heredoc format.
// With jsontypes.Normalized, the formatted JSON is preserved but semantic comparison prevents drift.
func TestAccConnectionResourceFilterJSONFormattingRaw(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	// jsontypes.Normalized preserves the original formatting but compares semantically
	expectedJSON := `{
  "data": {
    "attributes": {
      "payload": {
        "data": {
          "attributes": {
            "status": {
              "$or": [
                "completed",
                "failed",
                "approved",
                "declined",
                "needs_review"
              ]
            }
          }
        }
      }
    }
  }
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted("with_json_formatting_raw.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-json-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					// JSON preserves formatting
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.body.json", expectedJSON),
				),
			},
			// Re-apply should show no changes (no drift thanks to semantic comparison)
			{
				Config:   loadTestConfigFormatted("with_json_formatting_raw.tf", rName),
				PlanOnly: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Still has formatted JSON
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.body.json", expectedJSON),
				),
			},
		},
	})
}

// TestAccConnectionResourceFilterJSONFormatting tests pure jsonencode usage.
// This should work correctly without any issues.
func TestAccConnectionResourceFilterJSONFormatting(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_connection.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted("with_json_formatting.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-json-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.body.json", testFilterStatusJSON),
					resource.TestCheckResourceAttr(resourceName, "rules.0.filter_rule.headers.json",
						`{"x-api-version":"v1","x-webhook-type":"payment.status"}`),
				),
			},
		},
	})
}
