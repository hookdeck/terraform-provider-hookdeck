// Command teardown deletes what acceptance tests leave behind.
//
// It touches only objects whose name (label, for API keys) starts with
// "tf-acc-" and that are older than an hour, so a run in progress is never
// disturbed:
//
//   - connections, sources, destinations and transformations in the two
//     fixed test projects,
//   - test projects,
//   - test project API keys.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"terraform-provider-hookdeck/internal/sdkclient"
)

const (
	namePrefix   = "tf-acc-"
	projectNameA = "tf-ci-a"
	projectNameB = "tf-ci-b"
	projectHdr   = sdkclient.ProjectHeader
)

var baseURL = "https://api.hookdeck.com"

type resource struct {
	ID        string
	Name      string
	TeamID    string
	CreatedAt time.Time
}

// target is a set of deletable objects of one kind.
type target struct {
	Kind      string
	Path      string // collection path, deleted at Path + "/" + id
	ProjectID string // sent as the project header when not empty
	Items     []resource
}

type api struct {
	key    string
	client *http.Client
}

func (a api) do(method, path, projectID string, query url.Values) ([]byte, int, error) {
	u := fmt.Sprintf("%s/%s%s", baseURL, sdkclient.APIVersion, path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+a.key)
	req.Header.Set("Content-Type", "application/json")
	if projectID != "" {
		req.Header.Set(projectHdr, projectID)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

func parseResources(models []any) []resource {
	var out []resource
	for _, model := range models {
		m, ok := model.(map[string]any)
		if !ok {
			continue
		}
		var r resource
		r.ID, _ = m["id"].(string)
		// API keys have a label where everything else has a name.
		if name, ok := m["name"].(string); ok {
			r.Name = name
		} else if label, ok := m["label"].(string); ok {
			r.Name = label
		}
		r.TeamID, _ = m["team_id"].(string)
		if created, ok := m["created_at"].(string); ok {
			r.CreatedAt, _ = time.Parse(time.RFC3339, created)
		}
		out = append(out, r)
	}
	return out
}

// list returns every item of a collection. Collections answer with either
// a bare array or a page object with "models" and a "pagination.next"
// cursor.
func (a api) list(path, projectID string) ([]resource, error) {
	var all []resource
	cursor := ""
	for {
		query := url.Values{"limit": []string{"100"}}
		if cursor != "" {
			query.Set("next", cursor)
		}
		body, status, err := a.do(http.MethodGet, path, projectID, query)
		if err != nil {
			return nil, err
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("GET %s: API error (%d): %s", path, status, string(body))
		}
		var array []any
		if json.Unmarshal(body, &array) == nil {
			return append(all, parseResources(array)...), nil
		}
		var page struct {
			Models     []any `json:"models"`
			Pagination struct {
				Next string `json:"next"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		all = append(all, parseResources(page.Models)...)
		if page.Pagination.Next == "" {
			return all, nil
		}
		cursor = page.Pagination.Next
	}
}

// stale keeps the items a test left behind: named with the test prefix and
// created before cutoff.
func stale(items []resource, cutoff time.Time) []resource {
	var out []resource
	for _, r := range items {
		// A missing or unreadable created_at could be a running test's.
		if strings.HasPrefix(r.Name, namePrefix) && !r.CreatedAt.IsZero() && r.CreatedAt.Before(cutoff) {
			out = append(out, r)
		}
	}
	return out
}

func (a api) delete(t target, r resource) error {
	body, status, err := a.do(http.MethodDelete, t.Path+"/"+r.ID, t.ProjectID, nil)
	if err != nil {
		return err
	}
	if status >= 300 && !sdkclient.IsGoneStatus(status) {
		return fmt.Errorf("delete failed (%d): %s", status, string(body))
	}
	return nil
}

func printTarget(t target, dryRun bool) {
	label := t.Kind
	if !dryRun {
		fmt.Printf("🗑️  Deleting %d %s...\n", len(t.Items), label)
		return
	}
	fmt.Printf("\n📋 %s (%d):\n", label, len(t.Items))
	if len(t.Items) == 0 {
		fmt.Println("  None found")
		return
	}
	fmt.Println("  IDs to be deleted:")
	for _, r := range t.Items {
		fmt.Printf("    - %s (%s)\n", r.ID, r.Name)
	}
}

func confirmDeletion(total int) bool {
	fmt.Printf("\n⚠️  WARNING: This will delete %d objects permanently!\n", total)
	fmt.Print("Are you sure you want to proceed? (y/N): ")
	input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		_ = os.Setenv(strings.TrimSpace(parts[0]), value)
	}
	return nil
}

// collect finds everything to delete, in deletion order: the contents of the
// fixed projects (connections first, they reference the rest), then API
// keys, then projects.
func collect(a api, cutoff time.Time) ([]target, error) {
	projects, err := a.list("/projects", "")
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	var targets []target
	for _, p := range projects {
		if p.Name != projectNameA && p.Name != projectNameB {
			continue
		}
		for _, kind := range []string{"connections", "sources", "destinations", "transformations"} {
			items, err := a.list("/"+kind, p.ID)
			if err != nil {
				log.Printf("Error getting %s of %s: %v", kind, p.Name, err)
				continue
			}
			targets = append(targets, target{Kind: kind + " in " + p.Name, Path: "/" + kind, ProjectID: p.ID, Items: stale(items, cutoff)})
		}
	}
	keys, err := a.list("/organizations/current/api-keys", "")
	if err != nil {
		log.Printf("Error getting API keys: %v", err)
	} else {
		var projectKeys []resource
		for _, k := range keys {
			// Only project keys carry a team_id.
			if k.TeamID != "" {
				projectKeys = append(projectKeys, k)
			}
		}
		targets = append(targets, target{Kind: "project API keys", Path: "/organizations/current/api-keys", Items: stale(projectKeys, cutoff)})
	}
	targets = append(targets, target{Kind: "projects", Path: "/projects", Items: stale(projects, cutoff)})
	return targets, nil
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
	envFile := flag.String("env", ".env.test", "Path to environment file")
	autoApprove := flag.Bool("auto-approve", false, "Skip confirmation prompt (useful for CI/CD)")
	olderThan := flag.Duration("older-than", time.Hour, "Only delete objects created longer ago than this")
	help := flag.Bool("help", false, "Show usage information")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Hookdeck Teardown Tool\n======================\n\n")
		fmt.Fprintf(os.Stderr, "Deletes what acceptance tests leave behind when they fail or are interrupted.\n")
		fmt.Fprintf(os.Stderr, "Only objects named %q... and older than --older-than are deleted:\n", namePrefix)
		fmt.Fprintf(os.Stderr, "  - connections, sources, destinations and transformations in the projects %s and %s\n", projectNameA, projectNameB)
		fmt.Fprintf(os.Stderr, "  - projects\n  - project API keys (by label)\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  HOOKDECK_ORG_API_KEY  organization API key (required)\n")
		fmt.Fprintf(os.Stderr, "  HOOKDECK_API_BASE     API base URL (optional)\n")
	}
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if os.Getenv("HOOKDECK_ORG_API_KEY") == "" {
		if err := loadEnvFile(*envFile); err != nil {
			log.Printf("Warning: Could not load %s: %v", *envFile, err)
		}
	}
	apiKey := os.Getenv("HOOKDECK_ORG_API_KEY")
	if apiKey == "" {
		log.Fatal("HOOKDECK_ORG_API_KEY not found in environment or env file")
	}
	if base := os.Getenv("HOOKDECK_API_BASE"); base != "" {
		if !strings.HasPrefix(base, "http") {
			base = "https://" + base
		}
		baseURL = strings.TrimRight(base, "/")
	}

	if *dryRun {
		fmt.Println("🔍 DRY RUN MODE - Scanning Hookdeck Resources...")
	} else {
		fmt.Println("🧹 Hookdeck Resource Cleanup Tool")
	}
	fmt.Println(strings.Repeat("=", 51))

	a := api{key: apiKey, client: &http.Client{Timeout: 60 * time.Second}}
	targets, err := collect(a, time.Now().Add(-*olderThan))
	if err != nil {
		log.Fatal(err)
	}
	total := 0
	for _, t := range targets {
		printTarget(t, *dryRun)
		total += len(t.Items)
	}
	fmt.Println(strings.Repeat("=", 51))
	fmt.Printf("📊 Total objects: %d\n", total)

	if *dryRun {
		fmt.Println("\n💡 To delete these objects, run without --dry-run flag")
		return
	}
	if total == 0 {
		fmt.Println("\n✨ Nothing to delete.")
		return
	}
	if *autoApprove {
		fmt.Printf("\n🤖 Auto-approve enabled. Deleting %d objects...\n", total)
	} else if !confirmDeletion(total) {
		fmt.Println("❌ Deletion cancelled")
		return
	}

	fmt.Println("\n🚀 Starting deletion...")
	deleted := 0
	for _, t := range targets {
		if len(t.Items) == 0 {
			continue
		}
		fmt.Printf("\nDeleting %s...\n", t.Kind)
		for _, r := range t.Items {
			if err := a.delete(t, r); err != nil {
				log.Printf("  ❌ Failed to delete %s %s: %v", t.Kind, r.ID, err)
				continue
			}
			deleted++
			fmt.Printf("  ✅ Deleted %s (%s)\n", r.ID, r.Name)
		}
	}
	fmt.Println(strings.Repeat("=", 51))
	fmt.Printf("✨ Successfully deleted %d/%d objects\n", deleted, total)
}
