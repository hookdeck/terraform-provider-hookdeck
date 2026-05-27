package destination_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"terraform-provider-hookdeck/internal/provider"
	"terraform-provider-hookdeck/internal/sdkclient"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hookdeck": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HOOKDECK_API_KEY"); v == "" {
		t.Fatal("HOOKDECK_API_KEY must be set for acceptance tests")
	}
}

func loadTestConfigFormatted(t *testing.T, filename string, args ...interface{}) string {
	t.Helper()
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test fixture %q: %v", filename, err)
	}
	return fmt.Sprintf(string(content), args...)
}

// fetchDestinationConfig retrieves the destination's config object from the
// Hookdeck API. Used to verify the API state independently of Terraform state,
// because the `config` attribute is not Computed — Terraform state stores the
// user's input, not what the API actually has.
func fetchDestinationConfig(id string) (map[string]interface{}, error) {
	client := sdkclient.InitHookdeckSDKClient(
		os.Getenv("HOOKDECK_API_BASE"),
		os.Getenv("HOOKDECK_API_KEY"),
		"test",
	)
	resp, err := client.RawClient.SendRequest(context.Background(), "GET",
		fmt.Sprintf("/2025-07-01/destinations/%s", id),
		&sdkclient.RequestOptions{
			QueryParams: url.Values{"include": []string{"config.auth"}},
		})
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("GET destination returned %d: %s", resp.StatusCode, string(body))
	}
	var dest map[string]interface{}
	if err := json.Unmarshal(body, &dest); err != nil {
		return nil, err
	}
	config, ok := dest["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("destination response missing config object")
	}
	return config, nil
}

func checkAPIConfigValue(resourceName, key string, expected interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		config, err := fetchDestinationConfig(rs.Primary.ID)
		if err != nil {
			return err
		}
		got := config[key]
		switch want := expected.(type) {
		case nil:
			if got != nil {
				return fmt.Errorf("expected config.%s to be absent/null on API, got %v", key, got)
			}
		case float64:
			gotFloat, ok := got.(float64)
			if !ok || gotFloat != want {
				return fmt.Errorf("expected config.%s = %v on API, got %v", key, want, got)
			}
		case string:
			gotStr, ok := got.(string)
			if !ok || gotStr != want {
				return fmt.Errorf("expected config.%s = %q on API, got %v", key, want, got)
			}
		case bool:
			gotBool, ok := got.(bool)
			if !ok || gotBool != want {
				return fmt.Errorf("expected config.%s = %v on API, got %v", key, want, got)
			}
		default:
			return fmt.Errorf("unsupported expected type %T", expected)
		}
		return nil
	}
}

// TestAccDestinationResource_RemoveRateLimit verifies that removing
// `rate_limit` and `rate_limit_period` from the Terraform config actually
// clears them on the Hookdeck destination. The Hookdeck API merges config
// updates, so without explicit null-out logic, removed keys silently persist.
func TestAccDestinationResource_RemoveRateLimit(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_destination.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					checkAPIConfigValue(resourceName, "rate_limit", float64(10)),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// rate_limit being null is what disables rate limiting on the
					// destination. The API resets rate_limit_period to its default
					// ("second") when rate_limit is cleared, which is harmless.
					checkAPIConfigValue(resourceName, "rate_limit", nil),
				),
			},
		},
	})
}

// TestAccDestinationResource_AddRemoveReaddRateLimit verifies that adding,
// removing, and re-adding rate_limit works repeatedly without state
// corruption — exercising the diff logic across multiple Update cycles.
func TestAccDestinationResource_AddRemoveReaddRateLimit(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_destination.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Start minimal — no rate_limit.
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", nil),
				),
			},
			// Add rate_limit.
			{
				Config: loadTestConfigFormatted(t, "with_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", float64(10)),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
			// Remove rate_limit again.
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", nil),
				),
			},
			// Re-add rate_limit.
			{
				Config: loadTestConfigFormatted(t, "with_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", float64(10)),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
		},
	})
}

// TestAccDestinationResource_NoDriftAfterRateLimitRemoval verifies that
// re-applying the same config after removing rate_limit produces no plan
// diff. If our null-out logic incorrectly compared against the wrong baseline,
// every subsequent apply would try to re-null already-null keys.
func TestAccDestinationResource_NoDriftAfterRateLimitRemoval(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Apply with rate_limit set.
			{
				Config: loadTestConfigFormatted(t, "with_rate_limit.tf", rName),
			},
			// Remove rate_limit.
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
			},
			// Re-run the same config — expect empty plan.
			{
				Config:             loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccDestinationResource_RemoveHTTPMethod verifies the null-out fix
// generalizes beyond rate_limit. `http_method` is a top-level config field
// that documented to default to null (unlike `rate_limit_period`, which
// defaults to "second"), so removal should result in a true null on the API.
func TestAccDestinationResource_RemoveHTTPMethod(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_destination.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_http_method.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "http_method", "PUT"),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "http_method", nil),
				),
			},
		},
	})
}

// TestAccDestinationResource_RemovePathForwardingDisabledResetsToDefault
// verifies that removing the non-nullable `path_forwarding_disabled` boolean
// from a Terraform config resets the destination to the API's documented
// default (false). Booleans are non-nullable in the Hookdeck OpenAPI spec, so
// sending null would 422; destination/sdk.go looks up the documented default
// in nonNullableConfigDefaults and sends it explicitly so the field is reset,
// not retained.
func TestAccDestinationResource_RemovePathForwardingDisabledResetsToDefault(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_destination.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_path_forwarding_disabled.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "path_forwarding_disabled", true),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "without_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "path_forwarding_disabled", false),
				),
			},
		},
	})
}

// TestAccDestinationResource_RateLimitPeriodMergeBehavior probes whether
// the Hookdeck API auto-resets rate_limit_period to its default ("second")
// whenever rate_limit becomes null, or only when we explicitly send null
// for the period field.
//
// Scenario: start with both rate_limit and rate_limit_period set, then
// switch to a config that drops only rate_limit but keeps rate_limit_period.
// Our provider sends `{rate_limit: null, rate_limit_period: "concurrent"}`.
//
// If the period assertion passes — the API merges as expected and the
// "second" we see when both are removed is purely because we send null
// for both.
// If the period assertion fails (got "second") — the API auto-resets the
// period whenever rate_limit is null, regardless of what we send. That
// would be an API-side behavior, not a provider bug.
func TestAccDestinationResource_RateLimitPeriodMergeBehavior(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_destination.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", float64(10)),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "with_period_only.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "rate_limit", nil),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
		},
	})
}
