package connection_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"terraform-provider-hookdeck/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hookdeck": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("HOOKDECK_API_KEY"); v == "" {
		t.Fatal("HOOKDECK_API_KEY must be set for acceptance tests")
	}
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
			// Update with multiple rules
			{
				Config: loadTestConfigFormatted("with_multiple_rules.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-connection-multi-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
		},
	})
}

func loadTestConfigFormatted(filename string, rName string) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	// Format the template with the random suffix
	return fmt.Sprintf(string(content), rName)
}

func loadTestConfigFormattedWithUpdate(filename string, rName string, oldName string, newName string) string {
	config := loadTestConfigFormatted(filename, rName)
	// Replace the connection name for update tests
	return strings.ReplaceAll(config, oldName, newName)
}
