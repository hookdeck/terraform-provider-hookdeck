// Package acc is the harness shared by the acceptance tests.
//
// The tests need one credential, an organization API key in
// HOOKDECK_ORG_API_KEY. From it the harness finds (or creates) two fixed
// projects, ProjectNameA and ProjectNameB, and, for the tests that exercise
// project keys, mints one project key in the first project per test process.
//
// Every resource a test creates is named with NamePrefix so that
// cmd/teardown can tell test leftovers from anything else.
package acc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/provider"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	EnvAPIKey    = "HOOKDECK_API_KEY"
	EnvOrgAPIKey = "HOOKDECK_ORG_API_KEY"
	EnvProjectID = "HOOKDECK_PROJECT_ID"
	EnvAPIBase   = "HOOKDECK_API_BASE"

	// NamePrefix starts the name of everything a test creates.
	NamePrefix = "tf-acc-"

	// The fixed projects. Their names must not start with NamePrefix, so
	// that cleanup never deletes them.
	ProjectNameA = "tf-ci-a"
	ProjectNameB = "tf-ci-b"

	// MissingProjectID is a project ID that does not exist.
	MissingProjectID = "tm_doesnotexist"

	// FakeOrgAPIKey is a well-formed organization key that no test sends
	// to the API.
	FakeOrgAPIKey = projectscope.OrganizationKeyPrefix + "0000000000000000"
)

// ProviderFactories runs the provider under test in the test process.
// Each server gets its own provider instance, so tests can run in parallel.
var ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"hookdeck": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6WithError(provider.New("test")())()
	},
}

// Main is a package's TestMain. Credentials set for the provider in the
// environment are dropped, so that a test only sees the ones its PreCheck
// chose; keys minted during the run are deleted at the end.
func Main(m *testing.M) {
	_ = os.Unsetenv(EnvAPIKey)
	_ = os.Unsetenv(EnvProjectID)
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// SkipUnlessAcc skips before a test touches the API to build its steps.
func SkipUnlessAcc(t *testing.T) {
	t.Helper()
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("acceptance tests skipped unless %s is set", resource.EnvTfAcc)
	}
}

// UseAPIKey (with UseProject) is for the tests that run against a mock API:
// they set the process environment, so they cannot run in parallel. Tests
// against the real API pass credentials as Terraform variables instead; see
// Test.
//
// UseAPIKey makes the provider under test read apiKey from the environment
// for the rest of the test, so that keys never appear in a configuration.
func UseAPIKey(t *testing.T, apiKey string) {
	t.Helper()
	t.Setenv(EnvAPIKey, apiKey)
}

// UseProject sets the provider's project in the environment for the rest of
// the test. An empty projectID leaves the provider without one.
func UseProject(t *testing.T, projectID string) {
	t.Helper()
	t.Setenv(EnvProjectID, projectID)
}

// Creds are the provider credentials a test runs with. They reach the
// provider as Terraform variables, so tests using them can run in parallel.
type Creds struct {
	APIKey    string
	ProjectID string
}

// Org is the organization key with ProjectA as the provider's project
// (single-project mode). Most tests use it.
func Org(t *testing.T) Creds {
	t.Helper()
	return Creds{APIKey: OrgKey(t), ProjectID: ProjectA(t)}
}

// OrgNoProject is the organization key without a provider project, for
// tests that name projects themselves.
func OrgNoProject(t *testing.T) Creds {
	t.Helper()
	return Creds{APIKey: OrgKey(t)}
}

// ProjectKeyCreds is a project key for ProjectA.
func ProjectKeyCreds(t *testing.T) Creds {
	t.Helper()
	return Creds{APIKey: ProjectKey(t)}
}

// Variables returns the Terraform variables carrying the credentials.
func (c Creds) Variables() config.Variables {
	vars := config.Variables{VarAPIKey: config.StringVariable(c.APIKey)}
	if c.ProjectID != "" {
		vars[VarProjectID] = config.StringVariable(c.ProjectID)
	}
	return vars
}

const (
	VarAPIKey    = "hookdeck_api_key"
	VarProjectID = "hookdeck_project_id"

	variableDeclarations = `variable "hookdeck_api_key" {
  type      = string
  sensitive = true
  default   = null
}

variable "hookdeck_project_id" {
  type    = string
  default = null
}

`
	providerDeclaration = `provider "hookdeck" {
  api_key    = var.hookdeck_api_key
  project_id = var.hookdeck_project_id
}

`
)

var providerBlockPattern = regexp.MustCompile(`(?m)^\s*provider\s+"hookdeck"`)

// Test runs a test case against the real API as ProjectA with the
// organization key, in parallel with other tests.
func Test(t *testing.T, tc resource.TestCase) {
	t.Helper()
	TestAs(t, Org(t), tc)
}

// TestAs is Test with other credentials. The credentials are passed to the
// provider as variables of every step, unless a step sets its own
// ConfigVariables. A step's configuration gets the variable declarations,
// and a provider block unless it has one that reads var.hookdeck_api_key.
func TestAs(t *testing.T, c Creds, tc resource.TestCase) {
	t.Helper()
	tc.ProtoV6ProviderFactories = ProviderFactories
	steps := make([]resource.TestStep, len(tc.Steps))
	for i, step := range tc.Steps {
		if step.ConfigVariables == nil {
			step.ConfigVariables = c.Variables()
		}
		if step.Config != "" {
			header := variableDeclarations
			if !providerBlockPattern.MatchString(step.Config) {
				header += providerDeclaration
			}
			step.Config = header + step.Config
		}
		steps[i] = step
	}
	tc.Steps = steps
	resource.ParallelTest(t, tc)
}

// OrgKey returns the organization API key.
func OrgKey(t *testing.T) string {
	t.Helper()
	SkipUnlessAcc(t)
	key := os.Getenv(EnvOrgAPIKey)
	if key == "" {
		t.Fatalf("%s must be set for acceptance tests", EnvOrgAPIKey)
	}
	if !strings.HasPrefix(key, projectscope.OrganizationKeyPrefix) {
		t.Fatalf("%s must be an organization key (%s prefix)", EnvOrgAPIKey, projectscope.OrganizationKeyPrefix)
	}
	return key
}

func orgClient(t *testing.T) sdkclient.Client {
	t.Helper()
	return sdkclient.InitHookdeckSDKClient(os.Getenv(EnvAPIBase), OrgKey(t), "test")
}

// RandSuffix returns a random string for resource names.
func RandSuffix() string {
	return acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
}

type fixedProject struct {
	name string
	once sync.Once
	id   string
	err  error
}

var projectA, projectB = &fixedProject{name: ProjectNameA}, &fixedProject{name: ProjectNameB}

// ProjectA returns the ID of the first fixed project.
func ProjectA(t *testing.T) string { t.Helper(); return projectA.get(t) }

// ProjectB returns the ID of the second fixed project.
func ProjectB(t *testing.T) string { t.Helper(); return projectB.get(t) }

func (p *fixedProject) get(t *testing.T) string {
	t.Helper()
	client := orgClient(t)
	p.once.Do(func() { p.id, p.err = findOrCreateProject(context.Background(), client, p.name) })
	if p.err != nil {
		t.Fatalf("project %s: %v", p.name, p.err)
	}
	return p.id
}

type projectInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func projectsNamed(ctx context.Context, client sdkclient.Client, name string) ([]projectInfo, error) {
	var all []projectInfo
	if err := client.Do(ctx, http.MethodGet, "/projects", nil, nil, &all); err != nil {
		return nil, err
	}
	var matches []projectInfo
	for _, p := range all {
		if p.Name == name {
			matches = append(matches, p)
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].CreatedAt != matches[j].CreatedAt {
			return matches[i].CreatedAt < matches[j].CreatedAt
		}
		return matches[i].ID < matches[j].ID
	})
	return matches, nil
}

// findOrCreateProject looks the project up by name and creates it when it
// is missing. Test packages run as parallel processes, so on the very first
// run several may create it at once: the oldest wins and the others delete
// theirs.
func findOrCreateProject(ctx context.Context, client sdkclient.Client, name string) (string, error) {
	matches, err := projectsNamed(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("listing projects: %w", err)
	}
	if len(matches) > 0 {
		return matches[0].ID, nil
	}
	var created projectInfo
	if err := client.Do(ctx, http.MethodPost, "/projects", map[string]any{"name": name, "type": "event_gateway"}, nil, &created); err != nil {
		return "", fmt.Errorf("creating project: %w", err)
	}
	matches, err = projectsNamed(ctx, client, name)
	if err != nil {
		return "", fmt.Errorf("listing projects: %w", err)
	}
	if len(matches) > 1 && matches[0].ID != created.ID {
		_ = client.Do(ctx, http.MethodDelete, "/projects/"+created.ID, nil, nil, nil)
		return matches[0].ID, nil
	}
	return created.ID, nil
}

var (
	projectKeyOnce sync.Once
	projectKeyID   string
	projectKeyVal  string
	projectKeyErr  error
)

// ProjectKey returns a project API key for ProjectA, minted on first use
// and deleted when the test process ends.
func ProjectKey(t *testing.T) string {
	t.Helper()
	client := orgClient(t)
	project := ProjectA(t)
	projectKeyOnce.Do(func() {
		var created struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		projectKeyErr = client.Do(context.Background(), http.MethodPost, "/organizations/current/api-keys", map[string]any{
			"label":   NamePrefix + RandSuffix(),
			"type":    "project",
			"team_id": project,
		}, nil, &created)
		projectKeyID, projectKeyVal = created.ID, created.Key
		if projectKeyVal != "" && os.Getenv("GITHUB_ACTIONS") == "true" {
			fmt.Printf("::add-mask::%s\n", projectKeyVal)
		}
	})
	if projectKeyErr != nil {
		t.Fatalf("minting a project key: %v", projectKeyErr)
	}
	return projectKeyVal
}

func cleanup() {
	if projectKeyID == "" {
		return
	}
	client := sdkclient.InitHookdeckSDKClient(os.Getenv(EnvAPIBase), os.Getenv(EnvOrgAPIKey), "test")
	if err := client.Do(context.Background(), http.MethodDelete, "/organizations/current/api-keys/"+projectKeyID, nil, nil, nil); err != nil {
		fmt.Fprintf(os.Stderr, "deleting test project key %s: %v\n", projectKeyID, err)
	}
}

// Request calls the API with apiKey, in projectID when it is not empty. It
// returns the status and the decoded JSON object, if the body is one.
func Request(t *testing.T, apiKey, projectID, method, path string, payload any) (int, map[string]any) {
	t.Helper()
	status, body, err := Send(apiKey, projectID, method, path, nil, payload)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return status, body
}

// RequestCurrent is Request with the credentials of Org, for Check
// functions, which have no *testing.T. The project is resolved by the time
// a check runs.
func RequestCurrent(method, path string, query url.Values, payload any) (int, map[string]any, error) {
	return Send(os.Getenv(EnvOrgAPIKey), projectA.id, method, path, query, payload)
}

// Send calls the API under the API version, in projectID when it is not
// empty.
func Send(apiKey, projectID, method, path string, query url.Values, payload any) (int, map[string]any, error) {
	client := sdkclient.InitHookdeckSDKClient(os.Getenv(EnvAPIBase), apiKey, "test")
	opts := &sdkclient.RequestOptions{Headers: http.Header{}, QueryParams: query}
	if projectID != "" {
		opts.Headers.Set(sdkclient.ProjectHeader, projectID)
	}
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		opts.Body = bytes.NewReader(data)
		opts.Headers.Set("Content-Type", "application/json")
	}
	resp, err := client.RawClient.SendRequest(context.Background(), method, "/"+sdkclient.APIVersion+path, opts)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(body, &out)
	return resp.StatusCode, out, nil
}
