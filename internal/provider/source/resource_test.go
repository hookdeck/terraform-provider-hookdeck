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

func checkAPIConfigValue(resourceName, key string, expected interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		config, err := fetchSourceConfig(rs.Primary.ID)
		if err != nil {
			return err
		}
		got := config[key]
		switch want := expected.(type) {
		case nil:
			if got != nil {
				return fmt.Errorf("expected config.%s to be absent/null on API, got %v", key, got)
			}
		case []string:
			gotSlice, ok := got.([]interface{})
			if !ok {
				return fmt.Errorf("expected config.%s = %v on API, got %v (%T)", key, want, got, got)
			}
			if len(gotSlice) != len(want) {
				return fmt.Errorf("expected config.%s length %d, got %d (%v)", key, len(want), len(gotSlice), got)
			}
			for i, w := range want {
				gs, ok := gotSlice[i].(string)
				if !ok || gs != w {
					return fmt.Errorf("expected config.%s[%d] = %q on API, got %v", key, i, w, gotSlice[i])
				}
			}
		default:
			return fmt.Errorf("unsupported expected type %T", expected)
		}
		return nil
	}
}

// TestAccSourceResource_RemoveAllowedHTTPMethods verifies that removing
// `allowed_http_methods` from the Terraform config actually clears it on the
// Hookdeck source. The Hookdeck API merges config updates, so without explicit
// null-out logic, removed keys silently persist. Mirrors the destination
// rate_limit removal test, exercising the same fix in source/sdk.go.
func TestAccSourceResource_RemoveAllowedHTTPMethods(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_source.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: loadTestConfigFormatted("with_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					checkAPIConfigValue(resourceName, "allowed_http_methods", []string{"GET", "POST"}),
				),
			},
			{
				Config: loadTestConfigFormatted("without_allowed_http_methods.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAPIConfigValue(resourceName, "allowed_http_methods", nil),
				),
			},
		},
	})
}
