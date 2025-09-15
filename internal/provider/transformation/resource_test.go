package transformation_test

import (
	"fmt"
	"os"
	"path/filepath"
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

func loadTestConfigFormatted(filename string, args ...interface{}) string {
	path := filepath.Join("testdata", filename)
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	// Format the template with provided arguments
	return fmt.Sprintf(string(content), args...)
}

func TestAccTransformationResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_transformation.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: loadTestConfigFormatted("basic.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-transformation-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "code", "exports.handler = async (request, context) => { return request; };"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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
				Config: loadTestConfigFormatted("with_env.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-transformation-env-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "code", "exports.handler = async (request, context) => { return { ...request, apiKey: context.env.API_KEY }; };"),
					resource.TestCheckResourceAttrSet(resourceName, "env"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccTransformationResourceWithEnv(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourceName := fmt.Sprintf("hookdeck_transformation.test_%s", rName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with environment variables
			{
				Config: loadTestConfigFormatted("with_env.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-transformation-env-%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "env"),
					// Verify the env is properly set as JSON
					resource.TestCheckResourceAttr(resourceName, "env", fmt.Sprintf(`{"API_KEY":"test-key-%s","DEBUG":"true"}`, rName)),
				),
			},
			// Update to remove environment variables
			{
				Config: loadTestConfigFormatted("basic.tf", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("test-transformation-%s", rName)),
					resource.TestCheckNoResourceAttr(resourceName, "env"),
				),
			},
		},
	})
}
