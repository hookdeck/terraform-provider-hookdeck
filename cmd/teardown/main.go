package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const apiVersion = "2025-07-01"
const baseURL = "https://api.hookdeck.com"

type Resource struct {
	ID   string
	Name string
}

type ResourceList struct {
	Type  string
	Items []Resource
	Count int
}

func loadEnvFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove quotes if present
		value = strings.Trim(value, "\"'")
		os.Setenv(key, value)
	}

	return nil
}

func getResources(apiKey, resourceType string) ([]Resource, error) {
	var allResources []Resource
	nextCursor := ""

	for {
		// Build URL with pagination
		url := fmt.Sprintf("%s/%s/%s?limit=100", baseURL, apiVersion, resourceType)
		if nextCursor != "" {
			url += "&next=" + nextCursor
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		// Extract resources from models array
		if models, ok := result["models"].([]interface{}); ok {
			for _, model := range models {
				if m, ok := model.(map[string]interface{}); ok {
					resource := Resource{}
					if id, ok := m["id"].(string); ok {
						resource.ID = id
					}
					if name, ok := m["name"].(string); ok {
						resource.Name = name
					}
					allResources = append(allResources, resource)
				}
			}
		}

		// Check for pagination
		if pagination, ok := result["pagination"].(map[string]interface{}); ok {
			if next, ok := pagination["next"].(string); ok && next != "" {
				nextCursor = next
			} else {
				break // No more pages
			}
		} else {
			break // No pagination info
		}
	}

	return allResources, nil
}

func deleteResource(apiKey, resourceType, resourceID string) error {
	url := fmt.Sprintf("%s/%s/%s/%s", baseURL, apiVersion, resourceType, resourceID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func printResourceList(resources ResourceList, dryRun bool) {
	if dryRun {
		fmt.Printf("\n📋 %s (%d):\n", resources.Type, resources.Count)
		if resources.Count > 0 {
			fmt.Println("  IDs to be deleted:")
			for _, r := range resources.Items {
				if r.Name != "" {
					fmt.Printf("    - %s (%s)\n", r.ID, r.Name)
				} else {
					fmt.Printf("    - %s\n", r.ID)
				}
			}
		} else {
			fmt.Println("  None found")
		}
	} else {
		fmt.Printf("🗑️  Deleting %d %s...\n", resources.Count, resources.Type)
	}
}

func confirmDeletion(total int) bool {
	if total == 0 {
		fmt.Println("\n✨ No resources found to delete.")
		return false
	}

	fmt.Printf("\n⚠️  WARNING: This will delete %d resources permanently!\n", total)
	fmt.Print("Are you sure you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes"
}

func main() {
	// Parse flags
	dryRun := flag.Bool("dry-run", false, "Show what would be deleted without actually deleting")
	envFile := flag.String("env", ".env.test", "Path to environment file")
	autoApprove := flag.Bool("auto-approve", false, "Skip confirmation prompt (useful for CI/CD)")
	help := flag.Bool("help", false, "Show usage information")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Hookdeck Teardown Tool\n")
		fmt.Fprintf(os.Stderr, "======================\n\n")
		fmt.Fprintf(os.Stderr, "Deletes all Hookdeck resources (connections, sources, destinations, transformations)\n")
		fmt.Fprintf(os.Stderr, "from a workspace.\n\n")
		fmt.Fprintf(os.Stderr, "Useful for cleaning up after acceptance tests or Terraform test runs. Normally, tests\n")
		fmt.Fprintf(os.Stderr, "clean up after themselves, but if tests fail or are interrupted, resources may be left\n")
		fmt.Fprintf(os.Stderr, "behind. This tool helps clean up those orphaned resources.\n\n")
		fmt.Fprintf(os.Stderr, "Also helpful when 'terraform destroy' isn't sufficient (e.g., due to state inconsistencies,\n")
		fmt.Fprintf(os.Stderr, "provider bugs, or interrupted operations) - this tool directly removes all resources.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # See what would be deleted (dry run)\n")
		fmt.Fprintf(os.Stderr, "  %s --dry-run\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Delete all resources (with confirmation)\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Use a different environment file\n")
		fmt.Fprintf(os.Stderr, "  %s --env .env.production --dry-run\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Delete without confirmation (for CI/CD)\n")
		fmt.Fprintf(os.Stderr, "  %s --auto-approve\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Environment:\n")
		fmt.Fprintf(os.Stderr, "  HOOKDECK_API_KEY    API key for authentication (required)\n\n")
		fmt.Fprintf(os.Stderr, "Note: Resources are deleted in the following order:\n")
		fmt.Fprintf(os.Stderr, "  1. Connections (must be deleted before sources/destinations)\n")
		fmt.Fprintf(os.Stderr, "  2. Sources\n")
		fmt.Fprintf(os.Stderr, "  3. Destinations\n")
		fmt.Fprintf(os.Stderr, "  4. Transformations\n")
	}
	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Check if API key is already set in environment
	apiKey := os.Getenv("HOOKDECK_API_KEY")

	// Only load env file if API key not already set
	if apiKey == "" {
		if err := loadEnvFile(*envFile); err != nil {
			log.Printf("Warning: Could not load %s: %v", *envFile, err)
		}
		apiKey = os.Getenv("HOOKDECK_API_KEY")
	}

	// Final check for API key
	if apiKey == "" {
		log.Fatal("HOOKDECK_API_KEY not found in environment or env file")
	}

	if *dryRun {
		fmt.Println("🔍 DRY RUN MODE - Scanning Hookdeck Resources...")
	} else {
		fmt.Println("🧹 Hookdeck Resource Cleanup Tool")
	}
	fmt.Println("=" + strings.Repeat("=", 50))

	// Resource types in deletion order (connections first, then sources/destinations, then transformations)
	resourceTypes := []string{"connections", "sources", "destinations", "transformations"}
	allResources := make([]ResourceList, 0)
	totalCount := 0

	// Fetch all resources
	for _, resourceType := range resourceTypes {
		resources, err := getResources(apiKey, resourceType)
		if err != nil {
			log.Printf("Error getting %s: %v", resourceType, err)
			continue
		}

		list := ResourceList{
			Type:  resourceType,
			Items: resources,
			Count: len(resources),
		}
		allResources = append(allResources, list)
		totalCount += len(resources)
	}

	// Display results
	for _, resourceList := range allResources {
		printResourceList(resourceList, *dryRun)
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("📊 Total Resources: %d\n", totalCount)

	// If dry run, exit here
	if *dryRun {
		fmt.Println("\n💡 To delete these resources, run without --dry-run flag")
		return
	}

	// Ask for confirmation unless auto-approve is set
	if !*autoApprove {
		if !confirmDeletion(totalCount) {
			fmt.Println("❌ Deletion cancelled")
			return
		}
	} else {
		if totalCount == 0 {
			fmt.Println("\n✨ No resources found to delete.")
			return
		}
		fmt.Printf("\n🤖 Auto-approve enabled. Deleting %d resources...\n", totalCount)
	}

	// Perform deletion
	fmt.Println("\n🚀 Starting deletion...")
	deletedCount := 0

	for _, resourceList := range allResources {
		if resourceList.Count == 0 {
			continue
		}

		fmt.Printf("\nDeleting %s...\n", resourceList.Type)
		for _, resource := range resourceList.Items {
			if err := deleteResource(apiKey, resourceList.Type, resource.ID); err != nil {
				log.Printf("  ❌ Failed to delete %s %s: %v", resourceList.Type, resource.ID, err)
			} else {
				deletedCount++
				if resource.Name != "" {
					fmt.Printf("  ✅ Deleted %s (%s)\n", resource.ID, resource.Name)
				} else {
					fmt.Printf("  ✅ Deleted %s\n", resource.ID)
				}
			}
		}
	}

	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Printf("✨ Successfully deleted %d/%d resources\n", deletedCount, totalCount)
}
