package provider_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-hookdeck/internal/acc"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestMain(m *testing.M) { acc.Main(m) }

func loadFixture(t *testing.T, filename string, args ...interface{}) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read test fixture %q: %v", filename, err)
	}
	return fmt.Sprintf(string(content), args...)
}

// providerBlock returns a provider block with project_id. An empty
// projectID gives an empty string, leaving the provider to its defaults.
func providerBlock(projectID string) string {
	if projectID == "" {
		return ""
	}
	return fmt.Sprintf("provider \"hookdeck\" {\n  project_id = %q\n}\n", projectID)
}

// liveProviderBlock is providerBlock for tests against the real API, which
// pass their key as a variable.
func liveProviderBlock(projectID string) string {
	return fmt.Sprintf("provider \"hookdeck\" {\n  api_key    = var.hookdeck_api_key\n  project_id = %q\n}\n", projectID)
}

// sourceConfig returns a hookdeck_gateway_source named test. An empty
// projectID omits project_id.
func sourceConfig(t *testing.T, suffix, projectID string) string {
	t.Helper()
	if projectID == "" {
		return loadFixture(t, "source_without_project_id.tf", suffix)
	}
	return loadFixture(t, "source_with_project_id.tf", suffix, projectID)
}

// captureID stores the resource's ID for comparison in a later step.
func captureID(name string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		*into = rs.Primary.ID
		return nil
	}
}

func checkIDEquals(name string, want *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		if rs.Primary.ID != *want {
			return fmt.Errorf("%s: id changed from %q to %q", name, *want, rs.Primary.ID)
		}
		return nil
	}
}

func checkIDChanged(name string, old *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}
		if rs.Primary.ID == *old {
			return fmt.Errorf("%s: id %q did not change", name, *old)
		}
		return nil
	}
}

func importSourceWithProject(projectID string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[sourceAddr]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", sourceAddr)
		}
		return projectID + "/" + rs.Primary.ID, nil
	}
}

func importIDFromAttribute(name, attribute string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", name)
		}
		return rs.Primary.Attributes[attribute], nil
	}
}
