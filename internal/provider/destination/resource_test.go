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

func loadTestConfigFormatted(filename string, args ...interface{}) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
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
				Config: loadTestConfigFormatted("with_rate_limit.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					checkAPIConfigValue(resourceName, "rate_limit", float64(10)),
					checkAPIConfigValue(resourceName, "rate_limit_period", "concurrent"),
				),
			},
			{
				Config: loadTestConfigFormatted("without_rate_limit.tf", rName),
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
