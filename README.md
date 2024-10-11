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
      version = "~> 0.2"
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
  url  = "https://mock.hookdeck.com"
}

# Configure a connection
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
}
```

## Dependencies

This provider uses [Hookdeck API](https://hookdeck.com/docs/api) and [Hookdeck Go SDK](https://github.com/hookdeck/hookdeck-go-sdk) under the hood.

## Development

Enable pre-commit Git hooks to ensure any code changes are reflected in the documentation:

```sh
make enable-git-hooks
```
