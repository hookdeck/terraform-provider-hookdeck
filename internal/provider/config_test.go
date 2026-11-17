package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// releasedV3 is a released v3 provider, whose state holds the configuration
// as written.
const releasedV3 = "3.0.0-rc.1"

// A source and its authentication: Hookdeck returns the authentication in
// the source's config, which the source does not compare. Changing the
// authentication updates only the authentication.
func TestAccConfig_SourceWithSourceAuth(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:           loadFixture(t, "config_source_with_auth.tf", suffix, "first"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionCreate)}},
			},
			{
				Config:           loadFixture(t, "config_source_with_auth.tf", suffix, "first"),
				ConfigPlanChecks: emptyPlan(),
			},
			{
				Config: loadFixture(t, "config_source_with_auth.tf", suffix, "second"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionNoop),
					plancheck.ExpectResourceAction("hookdeck_gateway_source_auth.test", plancheck.ResourceActionUpdate),
				}},
			},
			{
				Config:           loadFixture(t, "config_source_with_auth.tf", suffix, "second"),
				ConfigPlanChecks: emptyPlan(),
			},
		},
	})
}

// The data sources return the configuration Hookdeck holds, with the
// defaults it adds.
func TestAccConfig_DataSources(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	has := func(want string) resource.CheckResourceAttrWithFunc {
		return func(value string) error {
			if !strings.Contains(value, want) {
				return fmt.Errorf("%s does not contain %s", value, want)
			}
			return nil
		}
	}

	acc.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: loadFixture(t, "config_datasources.tf", suffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_source.test", "config", has(`"allowed_http_methods":["POST"]`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_source.test", "config", has(`"custom_response"`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_destination.test", "config", has(`"url":"https://mock.hookdeck.com"`)),
					resource.TestCheckResourceAttrWith("data.hookdeck_gateway_destination.test", "config", has(`"path_forwarding_disabled"`)),
				),
			},
		},
	})
}

// State written by an earlier v3 provider holds the configuration as written.
// This version plans nothing for it, and still clears a key that is removed
// later.
func TestAccUpgrade_ReleasedV3State_Config(t *testing.T) {
	acc.SkipUnlessAcc(t)
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "hookdeck")
	suffix := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	acc.UseAPIKey(t, acc.ProjectKey(t))

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"hookdeck": {Source: "hookdeck/hookdeck", VersionConstraint: releasedV3},
				},
				Config: requiredProviders(releasedV3) + loadFixture(t, "v3_names_config.tf", suffix),
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + loadFixture(t, "v3_names_config.tf", suffix),
				ConfigPlanChecks:         emptyPlan(),
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + loadFixture(t, "v3_names_config_removed.tf", suffix),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(sourceAddr, plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction(destinationAddr, plancheck.ResourceActionUpdate),
				}},
				Check: func(s *terraform.State) error {
					id := s.RootModule().Resources[sourceAddr].Primary.ID
					status, body, err := acc.RequestCurrent("GET", "/sources/"+id, nil, nil)
					if err != nil || status > 299 {
						return fmt.Errorf("GET source %s: %d %v %v", id, status, body, err)
					}
					config, _ := body["config"].(map[string]any)
					if config["custom_response"] != nil {
						return fmt.Errorf("custom_response was not cleared: %v", config["custom_response"])
					}
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acc.ProviderFactories,
				Config:                   requiredProviders("") + loadFixture(t, "v3_names_config_removed.tf", suffix),
				ConfigPlanChecks:         emptyPlan(),
			},
		},
	})
}
