---
page_title: "Migrating from v2.x to v3.x"
description: "How to migrate from v2.x to v3.x of the Hookdeck Terraform Provider"
---

v3 of the Hookdeck Terraform Provider adds organization-level API keys, a `hookdeck_gateway_project` resource, and a naming convention that separates Event Gateway resources from Outpost resources. Existing v2 configurations keep working after a rename that Terraform performs without touching your Hookdeck resources.

Requires Terraform 1.8 or later.

## Summary of Changes

| Change | Action required |
|---|---|
| Event Gateway resources renamed with a `gateway_` prefix | Add `moved` blocks (no API calls) |
| `disabled_at` on destinations and connections is now read-only | Remove it from configuration if set |
| Organization API keys and `project_id` | Optional |
| Provider targets Hookdeck API version `2026-09-01` | None for existing resources |

## Resource Renames

Every Event Gateway resource and data source gains a `gateway_` prefix. Outpost resources, when added, use `outpost_`.

| v2 | v3 |
|---|---|
| `hookdeck_source` | `hookdeck_gateway_source` |
| `hookdeck_source_auth` | `hookdeck_gateway_source_auth` |
| `hookdeck_destination` | `hookdeck_gateway_destination` |
| `hookdeck_connection` | `hookdeck_gateway_connection` |
| `hookdeck_transformation` | `hookdeck_gateway_transformation` |
| `data.hookdeck_source` | `data.hookdeck_gateway_source` |
| `data.hookdeck_destination` | `data.hookdeck_gateway_destination` |
| `data.hookdeck_connection` | `data.hookdeck_gateway_connection` |

The provider implements resource state moves, so a `moved` block renames the resource in state without recreating it:

```hcl
# Before (v2)
resource "hookdeck_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
}

# After (v3)
resource "hookdeck_gateway_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
}

moved {
  from = hookdeck_source.stripe
  to   = hookdeck_gateway_source.stripe
}
```

`terraform plan` reports the move and no other change. Remove the `moved` blocks once every workspace has applied them.

Data sources have no state; rename them in place.

The v2 names remain available in v3 as deprecated aliases and emit a warning on every plan. They are removed in v4.


## `disabled_at` Is Read-Only

`disabled_at` on `hookdeck_gateway_destination` and `hookdeck_gateway_connection` was accepted in v2 configuration but never sent to Hookdeck. In v3 it is a computed attribute. Remove it from your configuration:

```hcl
# Before (v2)
resource "hookdeck_destination" "api" {
  name        = "api"
  disabled_at = "2025-01-01T00:00:00Z"   # had no effect
  # ...
}

# After (v3)
resource "hookdeck_gateway_destination" "api" {
  name = "api"
  # ...
}
```


## Organization API Keys and Projects

v2 authenticates with a project API key. v3 accepts either a project API key or an organization API key.

With a project API key nothing changes: every resource belongs to that key's project, and `project_id` must be omitted or equal to it.

With an organization API key, each resource needs a project. Set it once on the provider, or per resource:

```hcl
provider "hookdeck" {
  api_key    = var.hookdeck_org_api_key
  project_id = hookdeck_gateway_project.prod.id   # default for every resource
}

resource "hookdeck_gateway_project" "prod" {
  name = "prod"
}

resource "hookdeck_gateway_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
  # project_id inherited from the provider
}

resource "hookdeck_gateway_source" "stripe_staging" {
  project_id = hookdeck_gateway_project.staging.id   # overrides the provider default
  name       = "stripe"
  type       = "STRIPE"
}
```

Precedence is resource `project_id`, then provider `project_id`, then the key's own project. An organization key with no resolvable project fails at plan time.

`project_id` cannot change on an existing resource. Changing it forces replacement.

Existing projects are referenced with the `hookdeck_gateway_project` data source:

```hcl
data "hookdeck_gateway_project" "prod" {
  name = "prod"
}
```

Organization API keys are created in the Hookdeck dashboard under Organization Settings. Grant the key access to the projects Terraform manages, and `projects.write` if Terraform creates projects.

## Import IDs

Resources belonging to a project can be imported with the project prefixed:

```sh
terraform import hookdeck_gateway_source.stripe tm_xxx/src_xxx
```

The bare resource ID still works when the project resolves from the provider or the API key:

```sh
terraform import hookdeck_gateway_source.stripe src_xxx
```

Projects import by ID:

```sh
terraform import hookdeck_gateway_project.prod tm_xxx
```

## Resources Deleted Outside Terraform

In v2, a resource deleted from the Hookdeck dashboard caused `terraform plan` to fail. In v3 it is removed from state and the plan proposes recreating it. Hookdeck answers `404` for unknown ids and `410` for deleted resources; both are treated the same way.

## API Version

The provider targets Hookdeck API version `2026-09-01` for every resource. In v2, destinations used `2026-09-01` and other resources used `2025-07-01`. No resource shape visible to Terraform changed between those versions for sources, connections, transformations or source auth.

## Troubleshooting

### `Error: project_id is required`

The provider is configured with an organization API key and the resource has no `project_id`, and the provider has no default. Set one of them.

### `Error: project_id does not match the API key's project`

The provider is configured with a project API key and a resource sets a different `project_id`. Remove `project_id`, or switch to an organization API key.


### `moved` block reports "resource types are not compatible"

The provider version in the lock file predates v3. Run `terraform init -upgrade`.
