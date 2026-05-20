package source_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func loadTestConfigFormatted(filename string, args ...interface{}) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
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
				Config: loadTestConfigFormatted("with_custom_response.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					checkAPIConfigKeyIsObject(resourceName, "custom_response"),
				),
			},
			{
				Config: loadTestConfigFormatted("without_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigKeyAbsent(resourceName, "custom_response"),
				),
			},
		},
	})
}

// TestAccSourceResource_RemoveAllowedHTTPMethodsDoesNotError locks in that
// removing a non-nullable array-typed config field (`allowed_http_methods` is
// non-nullable per OpenAPI) does not cause the API to 422. The null-out logic
// in source/sdk.go skips arrays and booleans for this reason. The field
// retains its prior API value (merge-bug behavior preserved for non-nullable
// types) but the apply itself succeeds.
func TestAccSourceResource_RemoveAllowedHTTPMethodsDoesNotError(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted("with_allowed_http_methods.tf", rName),
			},
			{
				Config: loadTestConfigFormatted("without_allowed_http_methods.tf", rName),
			},
		},
	})
}
