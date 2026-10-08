---
page_title: "Hookdeck Provider"
description: |-
  Manage Hookdeck Event Gateway projects, sources, destinations, connections and transformations.
---

# Hookdeck Provider

The Hookdeck provider manages [Hookdeck Event Gateway](https://hookdeck.com) projects, sources, destinations, connections and transformations, and registers webhooks with third-party services.

Upgrading from v2? See the [v2 to v3 migration guide](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/v2-to-v3-migration).

## Example Usage

```terraform
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
  api_key = var.hookdeck_api_key
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}

resource "hookdeck_gateway_destination" "api" {
  name = "api"
  type = "HTTP"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
  })
}

resource "hookdeck_gateway_connection" "orders_to_api" {
  source_id      = hookdeck_gateway_source.orders.id
  destination_id = hookdeck_gateway_destination.api.id
}
```

## API Keys

The provider accepts a project API key or an organization API key, in `api_key` or the `HOOKDECK_API_KEY` environment variable. It tells them apart by the key's prefix.

| Key | Prefix | Reaches | Created in the dashboard under |
|---|---|---|---|
| Project API key | anything other than `hd_org_` | the one project it belongs to | Settings, Project, API Keys & Secrets |
| Organization API key | `hd_org_` | every project of the organization, or the projects selected on the key | Settings, Organization, API Keys |

Only an organization admin can create an organization API key. When creating one, choose "All projects" or "Specific projects", and give it the scopes your configuration needs (a `write` scope includes `read`):

| Scope | Needed for |
|---|---|
| `gateway.sources.write` | `hookdeck_gateway_source`, `hookdeck_gateway_source_auth` |
| `gateway.destinations.write` | `hookdeck_gateway_destination` |
| `gateway.connections.write` | `hookdeck_gateway_connection` |
| `gateway.transformations.write` | `hookdeck_gateway_transformation` |
| `projects.read` | the `hookdeck_gateway_project` data source |
| `projects.write` | the `hookdeck_gateway_project` resource, which also reads the project |

A project API key uses the same `gateway.*` scopes. It uses `projects.read` for the `hookdeck_gateway_project` data source and for the lookup of its own project at startup; the lookup is optional, as described under [Project API key](#project-api-key).

A key created with "Specific projects" also gets access to each project it creates.

## Projects

Every source, destination, connection, transformation and source auth belongs to a project. Each has a `project_id` attribute, which is always recorded in state, whether or not you set it. Project IDs start with `tm_`.

To find a project's ID:

- In Terraform, look the project up by name with the `hookdeck_gateway_project` data source and use its `id`:

  ```terraform
  data "hookdeck_gateway_project" "prod" {
    name = "prod"
  }
  ```

- With the [Hookdeck CLI](https://github.com/hookdeck/hookdeck-cli): `hookdeck project list --type gateway --output json`.
- With the API: `GET /projects` lists the projects the key can see.

A provider configuration works in one of two modes, single project or explicit:

| | Provider `project_id` set (attribute or `HOOKDECK_PROJECT_ID`) | Provider `project_id` not set |
|---|---|---|
| **Project API key** | Single project. Everything is in that project. | Single project. Everything is in the key's project. |
| **Organization API key** | Single project. Everything is in that project. | Explicit. Every resource and data source sets `project_id`. |

The provider `project_id` has to be known when Terraform plans: a literal, a variable or the environment variable. It cannot be an attribute of a resource or data source of the same provider configuration. To create a project and the resources in it in one configuration, leave the provider `project_id` unset and set `project_id` on each resource.

### Project API key

Everything is in the key's project. There is nothing to set. The provider asks the API once, at startup, which project the key belongs to. A key without the `projects.read` scope cannot answer that; the provider then carries on and takes each resource's project from the API's responses.

```terraform
variable "hookdeck_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_api_key # project API key
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}
```

### Project API key with `project_id`

The same as above, with the project set on the provider. The provider then skips the lookup of the key's project. If the key belongs to another project, the API refuses the first request and the provider reports that the key cannot access the project.

```terraform
variable "hookdeck_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key    = var.hookdeck_api_key # project API key
  project_id = "tm_prod"
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}
```

### Environment variables only

`HOOKDECK_PROJECT_ID` sets the provider `project_id`. The attribute wins when both are set. An organization API key with neither set is in explicit mode.

```terraform
# export HOOKDECK_API_KEY=...          a project or an organization API key
# export HOOKDECK_PROJECT_ID=tm_prod   optional with a project API key
provider "hookdeck" {}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}
```

### Organization API key, one project

Single-project mode with an organization API key. The provider sets `project_id` and resources leave it out. Setting the same `project_id` on a resource changes nothing; a different one is an error when planning.

```terraform
variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key    = var.hookdeck_org_api_key # hd_org_...
  project_id = "tm_prod"
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}

# Plan-time error "Project mismatch": this provider manages tm_prod only.
# resource "hookdeck_gateway_source" "other" {
#   project_id = "tm_staging"
#   name       = "orders"
#   type       = "HTTP"
# }
```

### Organization API key, several projects

Explicit mode. Every resource and data source sets `project_id`, so one configuration can create projects and the resources in them.

```terraform
variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_org_api_key # hd_org_...
}

resource "hookdeck_gateway_project" "staging" {
  name = "staging"
}

resource "hookdeck_gateway_source" "orders_staging" {
  project_id = hookdeck_gateway_project.staging.id
  name       = "orders"
  type       = "HTTP"
}

# An existing project, looked up by name.
data "hookdeck_gateway_project" "prod" {
  name = "prod"
}

resource "hookdeck_gateway_source" "orders_prod" {
  project_id = data.hookdeck_gateway_project.prod.id
  name       = "orders"
  type       = "HTTP"
}

# Plan-time error "Missing project_id": neither the resource nor the
# provider names a project.
# resource "hookdeck_gateway_source" "missing" {
#   name = "orders"
#   type = "HTTP"
# }
```

### One provider alias per project

Several projects with no `project_id` on any resource. Each provider alias sets one project and is in single-project mode. This also works with one project API key per alias.

```terraform
variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  alias      = "prod"
  api_key    = var.hookdeck_org_api_key
  project_id = "tm_prod"
}

provider "hookdeck" {
  alias      = "staging"
  api_key    = var.hookdeck_org_api_key
  project_id = "tm_staging"
}

resource "hookdeck_gateway_source" "orders_prod" {
  provider = hookdeck.prod
  name     = "orders"
  type     = "HTTP"
}

resource "hookdeck_gateway_source" "orders_staging" {
  provider = hookdeck.staging
  name     = "orders"
  type     = "HTTP"
}
```

### Rules

| Situation | Single project | Explicit |
|---|---|---|
| A resource or data source leaves `project_id` out | It is in the provider's project | Error when planning: `Missing project_id` |
| A resource sets `project_id` to the provider's project | No change. Adding it to or removing it from an existing resource also changes nothing | n/a |
| A resource sets `project_id` to another project | Error when planning: `Project mismatch` | It is in that project |
| `project_id` changes on an existing resource | n/a | The resource is replaced: deleted in the old project, created in the new one. A `hookdeck_gateway_source_auth` is updated in place to follow its replaced source |
| The provider `project_id` changes, with an organization API key | Every resource is replaced, except `hookdeck_gateway_source_auth`, which is updated in place to follow its replaced source | n/a |
| The project API key is swapped for another project's key, or the provider `project_id` changes with a project API key | Error when planning: `Project not reachable`. Nothing is changed | n/a |
| The key loses access to a resource's project | Error when planning, naming the project. The resource stays in state | Same as single project |
| Import ID | `<id>` or `<project_id>/<id>` | `<project_id>/<id>` |

Notes on replacement:

- A replaced source gets a new ID and a new URL.
- Before planning a replacement the provider sends one list request for that resource type to the new project. If the project does not exist, belongs to another organization, or the key has no access to it, the plan fails and nothing is deleted. The request needs only the scope the resource already needs.
- `hookdeck_gateway_source_auth` is part of its source. It moves with the source and is never replaced on its own.

### When the key cannot reach a resource's project

Reads, updates and deletes always go to the project recorded in state. If the provider's key cannot act on that project, the plan fails with `Project not reachable`, or with a message that starts with `The API key cannot access project`, and the resource stays in state. A resource is never silently created again in another project. Your options:

- Restore the API key or `project_id` that was changed by mistake.
- Stop managing the resource: `terraform state rm <address>`, or a `removed` block with `lifecycle { destroy = false }`. The resource stays in Hookdeck.
- Move the resource to the other project: use an organization API key with access to both projects. The plan then shows a replacement.

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_base` (String) Hookdeck API Base URL. Alternatively, can be configured using the `HOOKDECK_API_BASE` environment variable.
- `api_key` (String, Sensitive) Hookdeck API Key, either a project key or an organization key (`hd_org_` prefix). Alternatively, can be configured using the `HOOKDECK_API_KEY` environment variable.
- `project_id` (String) Project that every resource and data source of this provider configuration belongs to. Not needed with a project API key, which is bound to its project. With an organization API key, set it to manage a single project, or leave it unset and set `project_id` on each resource. Changing it replaces every resource. Alternatively, can be configured using the `HOOKDECK_PROJECT_ID` environment variable.
