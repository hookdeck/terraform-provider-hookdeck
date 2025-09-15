# Teardown

Deletes all Hookdeck resources (connections, sources, destinations, transformations) from a workspace. 

Useful for cleaning up after acceptance tests or Terraform test runs. Normally, tests clean up after themselves, but if tests fail or are interrupted, resources may be left behind. This tool helps clean up those orphaned resources.

Also helpful when `terraform destroy` isn't sufficient (e.g., due to state inconsistencies, provider bugs, or interrupted operations) - this tool directly removes all resources from the workspace.

## Usage

```bash
# Show help
go run ./cmd/teardown --help

# Dry run - see what would be deleted
go run ./cmd/teardown --dry-run

# Delete all resources (with confirmation prompt)
go run ./cmd/teardown

# Delete without confirmation (useful for CI/CD)
go run ./cmd/teardown --auto-approve

# Use a different environment file
go run ./cmd/teardown --env .env.production --dry-run
```

## Configuration

The tool uses the `HOOKDECK_API_KEY` environment variable for authentication.

By default, it loads from `.env.test`. You can override this:
- Using `--env` flag: `go run ./cmd/teardown --env .env.custom`
- Using environment variable directly: `HOOKDECK_API_KEY=your-key go run ./cmd/teardown`

## Deletion Order

Resources are deleted in this order to respect dependencies:
1. Connections (must be deleted before sources/destinations)
2. Sources
3. Destinations  
4. Transformations
