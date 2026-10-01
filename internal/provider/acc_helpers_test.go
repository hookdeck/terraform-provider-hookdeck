package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/provider"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	envAPIKey       = "HOOKDECK_API_KEY"
	envOrgAPIKey    = "HOOKDECK_ORG_API_KEY"
	envOrgProjectID = "HOOKDECK_ORG_PROJECT_ID"
	envProjectID    = "HOOKDECK_PROJECT_ID"

	// A well-formed organization key that no test sends to the API.
	fakeOrgAPIKey = projectscope.OrganizationKeyPrefix + "0000000000000000"
	// A project ID that does not exist.
	missingProjectID = "tm_doesnotexist"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hookdeck": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestMain(m *testing.M) {
	code := m.Run()
	deleteSecondOrgProject()
	os.Exit(code)
}

// skipUnlessAcc skips before a test touches the API to build its steps.
func skipUnlessAcc(t *testing.T) {
	t.Helper()
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("acceptance tests skipped unless %s is set", resource.EnvTfAcc)
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv(envAPIKey) == "" {
		t.Fatalf("%s must be set for acceptance tests", envAPIKey)
	}
	// A provider project set in the environment would change what "no
	// project_id" means in these tests.
	if os.Getenv(envProjectID) != "" {
		t.Fatalf("%s must not be set when running the provider test package", envProjectID)
	}
}

// skipUnlessOrg skips unless an organization API key and a project it can
// access are configured. The key needs projects.read, projects.write and
// the gateway scopes, in HOOKDECK_ORG_PROJECT_ID and in projects it creates.
func skipUnlessOrg(t *testing.T) {
	t.Helper()
	skipUnlessAcc(t)
	if os.Getenv(envOrgAPIKey) == "" || os.Getenv(envOrgProjectID) == "" {
		t.Skipf("%s and %s must be set for organization key tests", envOrgAPIKey, envOrgProjectID)
	}
	if !strings.HasPrefix(os.Getenv(envOrgAPIKey), projectscope.OrganizationKeyPrefix) {
		t.Fatalf("%s must be an organization key (%s prefix)", envOrgAPIKey, projectscope.OrganizationKeyPrefix)
	}
}

// useAPIKey makes the provider under test read apiKey from the environment
// for the rest of the test, so that keys never appear in a configuration.
func useAPIKey(t *testing.T, apiKey string) {
	t.Helper()
	t.Setenv(envAPIKey, apiKey)
}

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

// sourceConfig returns a hookdeck_gateway_source named test. An empty
// projectID omits project_id.
func sourceConfig(t *testing.T, suffix, projectID string) string {
	t.Helper()
	if projectID == "" {
		return loadFixture(t, "source_without_project_id.tf", suffix)
	}
	return loadFixture(t, "source_with_project_id.tf", suffix, projectID)
}

// apiRequest calls the API with apiKey, in projectID when it is not empty.
func apiRequest(t *testing.T, apiKey, projectID, method, path string, payload any) (int, map[string]interface{}) {
	t.Helper()
	client := sdkclient.InitHookdeckSDKClient(os.Getenv("HOOKDECK_API_BASE"), apiKey, "test")
	opts := &sdkclient.RequestOptions{Headers: http.Header{}}
	if projectID != "" {
		opts.Headers.Set(sdkclient.ProjectHeader, projectID)
	}
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		opts.Body = bytes.NewReader(data)
		opts.Headers.Set("Content-Type", "application/json")
	}
	resp, err := client.RawClient.SendRequest(context.Background(), method, "/"+sdkclient.APIVersion+path, opts)
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
// belongs to. Call skipUnlessAcc first.
func currentProjectID(t *testing.T) string {
	t.Helper()
	keyProjectOnce.Do(func() {
		client := sdkclient.InitHookdeckSDKClient(os.Getenv("HOOKDECK_API_BASE"), os.Getenv(envAPIKey), "test")
		id, err := client.KeyProjectID(context.Background())
		if err != nil {
			t.Fatalf("looking up the project of %s: %v", envAPIKey, err)
		}
		keyProjectID = id
	})
	if keyProjectID == "" {
		t.Fatal("could not determine the API key's project")
	}
	return keyProjectID
}

var (
	secondOrgProjectOnce sync.Once
	secondOrgProjectID   string
)

// secondOrgProject returns a second project in the organization of
// HOOKDECK_ORG_API_KEY, created once per test run and deleted in TestMain.
// Project creation is rate limited per organization, so tests share it.
func secondOrgProject(t *testing.T) string {
	t.Helper()
	secondOrgProjectOnce.Do(func() {
		name := "tf-acc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
		status, body := apiRequest(t, os.Getenv(envOrgAPIKey), "", "POST", "/projects", map[string]any{"name": name, "type": "event_gateway"})
		if status > 299 {
			t.Fatalf("creating project %s: %d %v", name, status, body)
		}
		secondOrgProjectID, _ = body["id"].(string)
	})
	if secondOrgProjectID == "" {
		t.Fatal("could not create a second project")
	}
	return secondOrgProjectID
}

func deleteSecondOrgProject() {
	if secondOrgProjectID == "" {
		return
	}
	client := sdkclient.InitHookdeckSDKClient(os.Getenv("HOOKDECK_API_BASE"), os.Getenv(envOrgAPIKey), "test")
	if err := client.Do(context.Background(), "DELETE", "/projects/"+secondOrgProjectID, nil, nil, nil); err != nil {
		fmt.Fprintf(os.Stderr, "deleting test project %s: %v\n", secondOrgProjectID, err)
	}
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
