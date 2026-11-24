---
page_title: "Managing Outpost with Terraform"
description: "Create and configure Hookdeck Outpost projects with the Hookdeck Terraform provider"
---

# Managing Outpost with Terraform

The `hookdeck_outpost_project` resource creates an [Outpost](https://hookdeck.com/docs/outpost) project and manages its configuration: topics, retry behavior, portal branding, observability and the other [Outpost settings](https://hookdeck.com/docs/outpost/self-hosting/configuration). That reference also covers self-hosted deployments: some of the settings it lists are managed by Hookdeck, and Hookdeck rejects them when they are set on a project. The `hookdeck_outpost_project` data source reads a project and its whole configuration.

Tenants and destinations are not managed by Terraform. Your application creates them through the Outpost API or SDKs.

## Requirements

Use an organization API key (`hd_org_` prefix) with the `projects.read` and `projects.write` scopes. An organization admin creates one in the dashboard; see [API Keys](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs#api-keys). A project API key cannot create projects. If the key is limited to specific projects, it gets access to the projects it creates.

```terraform
terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
      version = "~> 3.1"
    }
  }
}

variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_org_api_key
}
```

## Create a project

```terraform
resource "hookdeck_outpost_project" "deliveries" {
  name = "customer-deliveries"

  config = {
    TOPICS                   = "order.created,order.updated"
    RETRY_INTERVAL_SECONDS   = "60"
    MAX_RETRY_LIMIT          = "8"
    PORTAL_ORGANIZATION_NAME = "Acme"
    PORTAL_BRAND_COLOR       = "#3b82f6"
  }
}
```

Keys are Outpost environment variable names, and values are always strings. Applying waits until the project is ready, then sets the configuration.

Only the keys you set are managed. Every other key keeps its value, and a key you remove from the configuration returns to its default. The project restarts to apply a configuration change and may be briefly unavailable.

## Use the project

Your application talks to the project through the Outpost API at `https://api.outpost.hookdeck.com/2025-07-01`, or through an Outpost SDK, with an API key of the project. Terraform does not create that key. Create it in the dashboard, in the project's Settings under Secrets. The [quickstarts](https://hookdeck.com/docs/outpost/quickstarts/hookdeck-outpost-curl) show how to create a tenant and a destination and publish an event.

## Credentials

`sensitive_config` is the same as `config`, marked sensitive. Any key can go in either map, but not in both, so put the values you want hidden in `sensitive_config`. Terraform hides them in plan output and stores them in the state, so protect the state as you would for any secret. Pass the values in as sensitive variables rather than writing them in the configuration.

```terraform
variable "otlp_headers" {
  type      = string
  sensitive = true
}

resource "hookdeck_outpost_project" "deliveries" {
  name = "customer-deliveries"

  config = {
    OTEL_SERVICE_NAME           = "acme-outpost"
    OTEL_EXPORTER_OTLP_ENDPOINT = "https://otel.example.com"
  }

  sensitive_config = {
    OTEL_EXPORTER_OTLP_HEADERS = var.otlp_headers
  }
}
```

## Read the whole configuration

`config` and `sensitive_config` on the resource hold only the keys you manage. The data source returns the whole configuration of a project, including values you did not set.

```terraform
data "hookdeck_outpost_project" "deliveries" {
  name = "customer-deliveries"
}

output "retry_limit" {
  value = nonsensitive(data.hookdeck_outpost_project.deliveries.config["MAX_RETRY_LIMIT"])
}
```

The data source returns everything in one attribute, `config`. It is sensitive, because it can include credentials, and it is stored in the Terraform state like every other attribute. Wrap a value in `nonsensitive()` to output it, as above, only when it is not a secret.

## Import and deleting

The [`hookdeck_outpost_project` resource](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/outpost_project) page covers importing a project and what deleting one removes.
