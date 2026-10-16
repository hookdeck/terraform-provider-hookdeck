# Teardown

Deletes what acceptance tests leave behind when they fail or are interrupted. Normally tests clean up after themselves.

It deletes only objects whose name starts with `tf-acc-` (the prefix every test resource uses) and that were created more than an hour ago:

- connections, sources, destinations and transformations in the projects `tf-ci-a` and `tf-ci-b`
- projects
- project API keys, matched by label

Anything else, including the two fixed projects themselves, is left alone.

## Usage

```bash
# Show help
go run ./cmd/teardown --help

# Dry run - see what would be deleted
go run ./cmd/teardown --dry-run

# Delete with a confirmation prompt
go run ./cmd/teardown

# Delete without confirmation (useful for CI/CD)
go run ./cmd/teardown --auto-approve

# Change the age threshold (default 1h)
go run ./cmd/teardown --older-than 24h --dry-run

# Use a different environment file
go run ./cmd/teardown --env .env.production --dry-run
```

## Configuration

The tool uses the `HOOKDECK_ORG_API_KEY` environment variable for authentication, and `HOOKDECK_API_BASE` to target another environment.

By default, it loads from `.env.test` when the variable is not set. You can override this:
- Using `--env` flag: `go run ./cmd/teardown --env .env.custom`
- Using environment variable directly: `HOOKDECK_ORG_API_KEY=your-key go run ./cmd/teardown`

## Deletion Order

1. Connections (must be deleted before sources/destinations)
2. Sources
3. Destinations
4. Transformations
5. Project API keys
6. Projects
