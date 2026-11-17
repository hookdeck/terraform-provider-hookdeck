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

Your application talks to the project through the Outpost API at `https://api.outpost.hookdeck.com/2025-07-01`, or through an Outpost SDK, with an API key of the project. Terraform cannot create that key yet: create it in the dashboard, in the project's Settings under Secrets. The [quickstarts](https://hookdeck.com/docs/outpost/quickstarts/hookdeck-outpost-curl) show how to create a tenant and a destination and publish an event.

## Credentials

Keys whose values are secrets or can carry credentials go in `sensitive_config`, and so does a URL that includes a user or password, whatever its key. Terraform hides them in plan output and stores them in the state, so protect the state as you would for any secret. Pass the values in as sensitive variables rather than writing them in the configuration.

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

`terraform plan` fails when one of these keys, or such a URL, is in `config`. The page of the [`hookdeck_outpost_project` resource](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/outpost_project) lists them.

## Read the whole configuration

`config` on the resource holds only the keys you manage. The data source returns the whole configuration of a project, including values you did not set.

```terraform
data "hookdeck_outpost_project" "deliveries" {
  name = "customer-deliveries"
}

output "retry_limit" {
  value = data.hookdeck_outpost_project.deliveries.config["MAX_RETRY_LIMIT"]
}
```

Credentials and keys the provider does not know to be non-sensitive are in `sensitive_config`, which is stored in the Terraform state like every other attribute.

## Portal custom domain

Set `portal_custom_domain` to serve the project's portal on your own hostname. Applying adds the domain, and `portal_custom_domain_status` lists the DNS records to create for it.

```terraform
resource "hookdeck_outpost_project" "deliveries" {
  name                 = "customer-deliveries"
  portal_custom_domain = "portal.example.com"
}

output "portal_dns_records" {
  value = hookdeck_outpost_project.deliveries.portal_custom_domain_status.verification_records
}
```

Create the records at your DNS provider. They verify the hostname and issue its certificate. The status is `pending` until the hostname is verified, then `active`, which means the hostname is verified; the [resource page](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/outpost_project) describes the status fields and what changing or removing the hostname does.

## Several projects

Each Outpost project is a resource of its own, for example one per environment:

```terraform
resource "hookdeck_outpost_project" "staging" {
  name = "deliveries-staging"
}

resource "hookdeck_outpost_project" "production" {
  name = "deliveries-production"

  depends_on = [hookdeck_outpost_project.staging]
}
```

For now, projects are best created one after another, which `depends_on` does here.

## Import and deleting

The [`hookdeck_outpost_project` resource](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/resources/outpost_project) page covers importing a project and what deleting one removes.
