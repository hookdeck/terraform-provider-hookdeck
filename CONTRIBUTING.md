# Contributing

Thanks for your interest in contributing to the Hookdeck Terraform Provider. This guide covers setting up a development environment, running tests, and the release process.

## Prerequisites

- [Go](https://go.dev/dl/) — see [`go.mod`](go.mod) for the required version
- [Terraform CLI](https://developer.hashicorp.com/terraform/install) 1.9 or newer
- A [Hookdeck account](https://dashboard.hookdeck.com/signup) and API key for running acceptance tests

## Local development

### Build and install

```sh
go build
go install
```

### Use your local build with Terraform

Add a `dev_overrides` block to `~/.terraformrc` so Terraform uses your locally installed binary instead of fetching the published provider from the registry:

```hcl
provider_installation {
  dev_overrides {
    "hookdeck/hookdeck" = "/path/to/your/go/bin"
  }
  direct {}
}
```

See HashiCorp's [plugin framework tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install) for more detail.

### Documentation generation

Provider documentation under [`docs/`](docs/) is generated from the schema via [`terraform-plugin-docs`](https://github.com/hashicorp/terraform-plugin-docs). After making schema changes, regenerate it:

```sh
go generate ./...
```

CI verifies that the generated output is in sync with the schema. To regenerate automatically on every commit, enable the pre-commit hook (review [`.githooks/`](.githooks/) first):

```sh
make enable-git-hooks
```

### Linting

The project uses [`golangci-lint`](https://golangci-lint.run/). CI pins the version (currently `v2.5.0`); match it locally:

```sh
golangci-lint run
```

## Testing

### Unit tests

```sh
go test ./...
```

### Acceptance tests

Acceptance tests provision real resources on a Hookdeck workspace, run assertions, then destroy them. They require `HOOKDECK_API_KEY` and `TF_ACC=1`.

> [!WARNING]
> Use a dedicated test workspace, not your production workspace. Resources are created and destroyed on every run. If tests are interrupted, orphaned resources can be cleaned up with [`cmd/teardown`](cmd/teardown/README.md).

Set up your environment once:

```sh
cp .env.test.example .env.test
# edit .env.test and set HOOKDECK_API_KEY
```

Run the full suite:

```sh
make testacc
```

Target a single package:

```sh
TEST=./internal/provider/source make testacc
```

Run a specific test:

```sh
TEST=./internal/provider/destination RUN=TestAccDestinationResource_RemoveRateLimit make testacc
```

Point at a different environment (e.g. staging):

```sh
HOOKDECK_API_BASE=https://api.staging.hookdeck.com make testacc
```

## Continuous integration

The [`Tests`](.github/workflows/test.yml) workflow runs on every pull request and push to `main`:

- Build, lint, and `go generate` sync check on all PRs.
- Acceptance tests across a Terraform CLI version matrix, on same-repository PRs and pushes to `main`.
- Fork PRs and Dependabot PRs **skip the acceptance job** because they cannot access `HOOKDECK_API_KEY` (GitHub policy for forks; restricted token scope for Dependabot). They are validated by the acceptance run that follows merge to `main`.
- Acceptance runs are serialized via a `concurrency` group to stay within Hookdeck's 240 req/min API limit.

## Submitting changes

1. Fork the repository and create a topic branch from `main`.
2. Make your change, including tests where applicable.
3. Run `go generate ./...`, `golangci-lint run`, and the relevant tests locally.
4. Open a pull request describing the change and linking any related issue.
5. The maintainers will review; reviewers from the Hookdeck team will rerun acceptance tests if the PR is from a fork.

## Release process

Releases are tag-driven. Pushing a `v*` tag triggers the [`Release`](.github/workflows/release.yml) workflow, which uses [GoReleaser](https://goreleaser.com) to build, sign, and publish binaries to a new GitHub release. The Terraform Registry picks the release up automatically.

To cut a release:

1. Confirm `main` is green.
2. Decide on the version per [SemVer](https://semver.org).
3. Create a GitHub release with the version as both the title and the tag (for example `v2.3.0`), targeting `main`. GoReleaser will attach the signed binaries.
4. Once the workflow completes, edit the release notes to summarize changes (use prior releases as a template).
