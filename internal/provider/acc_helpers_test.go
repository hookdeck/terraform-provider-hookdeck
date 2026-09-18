package provider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"terraform-provider-hookdeck/internal/provider"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	envAPIKey       = "HOOKDECK_API_KEY"
	envOrgAPIKey    = "HOOKDECK_ORG_API_KEY"
	envOrgProjectID = "HOOKDECK_ORG_PROJECT_ID"
	envProjectID    = "HOOKDECK_PROJECT_ID"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hookdeck": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv(envAPIKey) == "" {
		t.Fatalf("%s must be set for acceptance tests", envAPIKey)
	}
	// A provider-level default project set in the environment would change
	// what "no project_id" means in these tests.
	if os.Getenv(envProjectID) != "" {
		t.Fatalf("%s must not be set when running the provider test package", envProjectID)
	}
}

// testAccOrgPreCheck skips unless an organization API key with access to a
// project is configured. The key needs projects.write plus a grant on
// HOOKDECK_ORG_PROJECT_ID.
func testAccOrgPreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	if os.Getenv(envOrgAPIKey) == "" || os.Getenv(envOrgProjectID) == "" {
		t.Skipf("%s and %s must be set for organization key tests", envOrgAPIKey, envOrgProjectID)
	}
	if !strings.HasPrefix(os.Getenv(envOrgAPIKey), "hd_org_") {
		t.Fatalf("%s must be an organization key (hd_org_ prefix)", envOrgAPIKey)
	}
}

func loadFixture(t *testing.T, filename string, args ...interface{}) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read test fixture %q: %v", filename, err)
	}
	return fmt.Sprintf(string(content), args...)
}

// orgProvider returns a provider block using the organization key. An empty
// projectID omits the provider default.
func orgProvider(projectID string) string {
	var sb strings.Builder
	sb.WriteString("provider \"hookdeck\" {\n")
	sb.WriteString(fmt.Sprintf("  api_key = %q\n", os.Getenv(envOrgAPIKey)))
	if projectID != "" {
		sb.WriteString(fmt.Sprintf("  project_id = %q\n", projectID))
	}
	sb.WriteString("}\n")
	return sb.String()
}

func rawClient(apiKey string) sdkclient.Client {
	return sdkclient.InitHookdeckSDKClient(os.Getenv("HOOKDECK_API_BASE"), apiKey, "test")
}

func apiRequest(t *testing.T, apiKey, method, path string) (int, map[string]interface{}) {
	t.Helper()
	resp, err := rawClient(apiKey).RawClient.SendRequest(context.Background(), method, path, nil)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out map[string]interface{}
	_ = json.Unmarshal(body, &out)
	return resp.StatusCode, out
}

var (
	keyProjectOnce sync.Once
	keyProjectID   string
)

// currentProjectID returns the project the HOOKDECK_API_KEY project key
// belongs to.
func currentProjectID(t *testing.T) string {
	t.Helper()
	keyProjectOnce.Do(func() {
		resp, err := rawClient(os.Getenv(envAPIKey)).RawClient.SendRequest(context.Background(), "GET", "/2026-09-01/projects", nil)
		if err != nil {
			t.Fatalf("GET /projects: %v", err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var projects []map[string]interface{}
		if err := json.Unmarshal(body, &projects); err != nil || len(projects) != 1 {
			t.Fatalf("GET /projects with a project key should return exactly one project, got %d: %s", resp.StatusCode, body)
		}
		keyProjectID, _ = projects[0]["id"].(string)
	})
	if keyProjectID == "" {
		t.Fatal("could not determine the API key's project")
	}
	return keyProjectID
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
