# Hookdeck Terraform Provider

_The [Hookdeck Event Gateway](https://hookdeck.com) enables engineering teams to build, deploy, observe, and scale event-driven applications. For more information, see the [Hookdeck documentation](https://hookdeck.com/docs)._

The Hookdeck Terraform provider enables you to manage your Hookdeck projects using IaC (Infrastructure-as-Code), including managing your sources, destinations, connections, transformations, and more. It also supports webhook registration workflow that allows you to configure webhooks as part of your CI/CD processes.

## Installation

To install Hookdeck Terraform provider:

1. Obtain a project API key from [the dashboard](https://dashboard.hookdeck.com/settings/project/secrets). To manage several projects or to create projects, use an organization API key instead: see [API Keys](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs#api-keys) in the provider documentation.
2. Add the following to your Terraform configuration file:

```hcl
terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
      version = "~> 3.0"
    }
  }
}

variable "hookdeck_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  # set HOOKDECK_API_KEY env var or optionally specify the key in the provider configuration
  api_key = var.hookdeck_api_key
}
```

Upgrading from v2? See the [v2 to v3 migration guide](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/v2-to-v3-migration). Renaming existing resources to the v3 names with `moved` blocks needs Terraform 1.8 or OpenTofu 1.10, or later.

## Using the provider

This README gives a basic example; for more examples, see the [examples/](examples/) folder, the rendered documentation on the [Terraform Registry](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs), or [docs folder](docs/) in this repository.

```hcl
# Configure a source
resource "hookdeck_gateway_source" "my_source" {
  name = "my_source"
}

# Configure a destination
resource "hookdeck_gateway_destination" "my_destination" {
  name = "my_destination"
  type = "HTTP"
  config = jsonencode({
    url = "https://myapp.example.com/api"
  })
}

# Configure a connection
resource "hookdeck_gateway_connection" "my_connection" {
  source_id      = hookdeck_gateway_source.my_source.id
  destination_id = hookdeck_gateway_destination.my_destination.id
}
```

For [Source `config`](https://hookdeck.com/docs/api#source-object) and [Destination `config`](https://hookdeck.com/docs/api#destination-object) you must provide a JSON object. This means you do not get validation on the `config` property within your IDE or when running `terraform plan`. However, when running `terraform apply` the Hookdeck API will provide error responses if invalid configuration is received.

## Dependencies

This provider is built on top of the [Hookdeck API](https://hookdeck.com/docs/api).

## Contributing

Contributions are welcome. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development setup, running acceptance tests, documentation generation, and the release process. If you are working on this repository with an AI coding agent, see [`AGENTS.md`](AGENTS.md).
