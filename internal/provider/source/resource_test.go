package source_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
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

// fetchSourceConfig retrieves the source's config object from the Hookdeck
// API. Used to verify the API state independently of Terraform state, because
// the `config` attribute is not Computed — Terraform state stores the user's
// input, not what the API actually has.
func fetchSourceConfig(id string) (map[string]interface{}, error) {
	client := sdkclient.InitHookdeckSDKClient(
		os.Getenv("HOOKDECK_API_BASE"),
		os.Getenv("HOOKDECK_API_KEY"),
		"test",
	)
	resp, err := client.RawClient.SendRequest(context.Background(), "GET",
		fmt.Sprintf("/2025-07-01/sources/%s", id), nil)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("GET source returned %d: %s", resp.StatusCode, string(body))
	}
	var src map[string]interface{}
	if err := json.Unmarshal(body, &src); err != nil {
		return nil, err
	}
	config, ok := src["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("source response missing config object")
	}
	return config, nil
}

func checkAPIConfigKeyAbsent(resourceName, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		config, err := fetchSourceConfig(rs.Primary.ID)
		if err != nil {
			return err
		}
		if got := config[key]; got != nil {
			return fmt.Errorf("expected config.%s to be absent/null on API, got %v", key, got)
		}
		return nil
	}
}

func checkAPIConfigKeyIsObject(resourceName, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		config, err := fetchSourceConfig(rs.Primary.ID)
		if err != nil {
			return err
		}
		if _, ok := config[key].(map[string]interface{}); !ok {
			return fmt.Errorf("expected config.%s to be an object on API, got %v", key, config[key])
		}
		return nil
	}
}

func checkAPIConfigStringSlice(resourceName, key string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		config, err := fetchSourceConfig(rs.Primary.ID)
		if err != nil {
			return err
		}
		raw, ok := config[key].([]interface{})
		if !ok {
			return fmt.Errorf("expected config.%s to be an array on API, got %v (%T)", key, config[key], config[key])
		}
		if len(raw) != len(expected) {
			return fmt.Errorf("expected config.%s length %d, got %d (%v)", key, len(expected), len(raw), raw)
		}
		got := make(map[string]bool, len(raw))
		for _, v := range raw {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("expected config.%s to be []string, got element %v (%T)", key, v, v)
			}
			got[s] = true
		}
		for _, w := range expected {
			if !got[w] {
				return fmt.Errorf("expected config.%s to contain %q on API, got %v", key, w, raw)
			}
		}
		return nil
	}
}

// TestAccSourceResource_RemoveCustomResponse mirrors the destination
// rate_limit removal test for sources: applies a source with a
// `custom_response` object set, removes it, and verifies via direct API read
// that the key is cleared. `custom_response` is nullable in the OpenAPI spec
// so the provider's null-out logic in source/sdk.go applies cleanly.
func TestAccSourceResource_RemoveCustomResponse(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_source.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_custom_response.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					checkAPIConfigKeyIsObject(resourceName, "custom_response"),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "without_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigKeyAbsent(resourceName, "custom_response"),
				),
			},
		},
	})
}

// TestAccSourceResource_RemoveAllowedHTTPMethodsResetsToDefault verifies that
// removing the non-nullable `allowed_http_methods` field from a Terraform
// config resets the source to the API's documented default
// (["PUT","POST","PATCH","DELETE"]). Sending null would 422; source/sdk.go
// looks up the documented default in nonNullableConfigDefaults and sends it
// explicitly so the field is reset, not retained.
func TestAccSourceResource_RemoveAllowedHTTPMethodsResetsToDefault(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_source.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigStringSlice(resourceName, "allowed_http_methods", []string{"GET", "POST"}),
				),
			},
			{
				Config: loadTestConfigFormatted(t, "without_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigStringSlice(resourceName, "allowed_http_methods",
						[]string{"PUT", "POST", "PATCH", "DELETE"}),
				),
			},
		},
	})
}

// TestAccSourceResource_NestedConfigObjectsValidatedWhole locks in the
// invariant that justifies why source/sdk.go's toUpdatePayload only diffs
// top-level config keys: the Hookdeck API rejects any update where a nested
// config object is sent with a required inner field missing. Because the
// API never accepts a partial nested object, the provider has no
// user-input shape that would require diffing inner keys to clear them.
//
// Step 1 applies a valid HMAC `auth` block. Step 2 attempts to update with
// `header_key` removed and expects a 422 from the API. If this test ever
// starts passing step 2 (i.e. the API loosens validation and accepts
// partial nested updates), the top-level-only diff in toUpdatePayload may
// need to recurse, and this test failure is the trigger to revisit.
//
// Empirical basis: also confirmed against `custom_response`, which the
// OpenAPI spec marks `content_type` and `body` as both required and the
// API rejects identically. HMAC `auth` is used here because the spec
// declares no `required` inner fields yet the backend still enforces
// them, which makes this the strongest case to lock in.
func TestAccSourceResource_NestedConfigObjectsValidatedWhole(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted(t, "with_hmac_auth.tf", rName),
			},
			{
				Config:      loadTestConfigFormatted(t, "with_hmac_auth_no_header_key.tf", rName),
				ExpectError: regexp.MustCompile(`config\.auth\.header_key\s+is required`),
			},
		},
	})
}
