# Hookdeck Terraform Provider

_The [Hookdeck Event Gateway](https://hookdeck.com) enables engineering teams to build, deploy, observe, and scale event-driven applications. For more information, see the [Hookdeck documentation](https://hookdeck.com/docs)._

The Hookdeck Terraform provider enables you to manage your Hookdeck workspaces using IaC (Infrastructure-as-Code), including managing your sources, destinations, connections, transformations, and more. It also supports webhook registration workflow that allows you to configure webhooks as part of your CI/CD processes.

## Installation

To install Hookdeck Terraform provider:

1. Obtain your Hookdeck API key from [the dashboard](https://dashboard.hookdeck.com/workspace/secrets)
2. Add the following to your Terraform configuration file:

```hcl
terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
    }
  }
}

provider "hookdeck" {
  # set HOOKDECK_API_KEY env var or optionally specify the key in the provider configuration
  api_key = var.hookdeck_api_key
}
```

## Using the provider

This README gives a basic example; for more examples, see the [examples/](examples/) folder, the rendered documentation on the [Terraform Registry](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs), or [docs folder](docs/) in this repository.

```hcl
# Configure a source
resource "hookdeck_source" "my_source" {
  name = "my_source"
}

# Configure a destination
resource "hookdeck_destination" "my_destination" {
  name = "my_destination"
  type = "HTTP"
  config = jsonencode({
    url  = "https://myapp.example.com/api"
  })
}

# Configure a connection
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
}
```

For [Source `config`](https://hookdeck.com/docs/api#source-object) and [Destination `config`](https://hookdeck.com/docs/api#destination-object) you must provide a JSON object. This means you do not get validation on the `config` property within your IDE or when running `terraform plan`. However, when running `terraform apply` the Hookdeck API will provide error responses if invalid configuration is received.

## Dependencies

This provider is built on top of the [Hookdeck API](https://hookdeck.com/docs/api).

## Development

### Running locally

See https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install

### Brief details

Build and install:

```
go build
go install
```

Override the provider in a `~/.terraformrc`:

```
provider_installation {

  dev_overrides {
      "hookdeck/hookdeck" = "/Users/leggetter/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```


### Release

Released are managed via [GitHub Releases](https://github.com/hookdeck/terraform-provider-hookdeck/releases).

To release, create a new release with a name representing the SemVer version. Also, create a tag with the same version. A GitHub action is triggered via the new Tag creation and uses [GoReleaser](https://goreleaser.com) to create a new set of release assets for the Hookdeck Terraform Provider.

### Notes

Enable pre-commit Git hooks to ensure any code changes are reflected in the documentation:

```sh
make enable-git-hooks
```
