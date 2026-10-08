---
page_title: "Migrating from v2.x to v3.x"
description: "How to migrate from v2.x to v3.x of the Hookdeck Terraform Provider"
---

v3 of the Hookdeck Terraform Provider adds organization API keys, a `hookdeck_gateway_project` resource, and a `gateway_` prefix on Event Gateway resources. Outpost resources will use `outpost_`.

A v2 configuration keeps working with v3 as it is, unless it sets `disabled_at`. The v2 resource names are deprecated, and renaming them does not recreate anything in Hookdeck.

## Summary of Changes

| Change | Action required |
|---|---|
| Event Gateway resources renamed with a `gateway_` prefix | Rename, with `moved` blocks. Can be done later: the v2 names work until v4 |
| Every resource records its `project_id` | None. Recorded in state on the first apply |
| The new names have `project_id` and no `team_id` | When renaming, change references to `.team_id` into `.project_id` |
| `disabled_at` on destinations and connections is read-only | Remove it from configuration if set |
| Organization API keys, `project_id` on the provider and on resources | Optional |
| New `hookdeck_gateway_project` resource and data source | Optional |
| A resource deleted outside Terraform no longer fails the plan | None |
| Every resource uses Hookdeck API version `2026-09-01` | None |

## Requirements

| What | Terraform | OpenTofu |
|---|---|---|
| Using v3, with the v2 names or the new names | 1.0 or later | 1.6 or later |
| Renaming existing resources with `moved` blocks | 1.8 or later | 1.10 or later |

Earlier versions cannot move a resource to a different resource type. On those versions, keep the v2 names or rename with [`state rm` and `import`](#renaming-without-moved-blocks).

## Upgrade Steps

1. Set the provider version constraint to `~> 3.0` and run `terraform init -upgrade`.
2. If your configuration sets `disabled_at`, [remove it](#disabled_at-is-read-only).
3. Run `terraform plan`. It reports no changes, plus a `Deprecated` warning for each resource and data source block that uses a v2 name. Terraform before 1.12 and OpenTofu show each of these warnings twice on `plan`, so 6 blocks give 12 warnings. Terraform 1.15 and later also warn with `Deprecated value used` wherever a v2-named resource is referenced.
4. Run `terraform apply`. Nothing changes in Hookdeck; the apply records each resource's `project_id` in state.
5. Rename the resources, now or any time before v4: see [Resource Renames](#resource-renames).

The version constraint for step 1:

```terraform
terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
      version = "~> 3.0"
    }
  }
}
```

## Resource Renames

Every Event Gateway resource and data source gains a `gateway_` prefix.

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

`hookdeck_webhook_registration` keeps its name: it registers webhooks with third-party services and is not specific to Event Gateway.

The v2 names remain in v3 as deprecated aliases and give a warning on every plan. They are removed in v4.

To rename a resource, change its type, update every reference to it, and add a `moved` block. Without the `moved` block Terraform plans to destroy the resource and create a new one, and a new source has a new URL.

The new names have no `team_id` attribute: it held the same value as `project_id`. A reference such as `hookdeck_source.stripe.team_id` becomes `hookdeck_gateway_source.stripe.project_id`. The v2 names keep `team_id` until v4.

Before:

```terraform
resource "hookdeck_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
}

resource "hookdeck_destination" "api" {
  name = "api"
  type = "HTTP"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
  })
}

resource "hookdeck_connection" "stripe_to_api" {
  source_id      = hookdeck_source.stripe.id
  destination_id = hookdeck_destination.api.id
}
```

After:

```terraform
resource "hookdeck_gateway_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
}

resource "hookdeck_gateway_destination" "api" {
  name = "api"
  type = "HTTP"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
  })
}

resource "hookdeck_gateway_connection" "stripe_to_api" {
  source_id      = hookdeck_gateway_source.stripe.id
  destination_id = hookdeck_gateway_destination.api.id
}

moved {
  from = hookdeck_source.stripe
  to   = hookdeck_gateway_source.stripe
}

moved {
  from = hookdeck_destination.api
  to   = hookdeck_gateway_destination.api
}

moved {
  from = hookdeck_connection.stripe_to_api
  to   = hookdeck_gateway_connection.stripe_to_api
}
```

`terraform plan` lists each resource as "has moved to" and ends with `Plan: 0 to add, 0 to change, 0 to destroy`. Apply it. Remove the `moved` blocks once every state that uses the configuration has applied them.

Data sources need no `moved` block. Rename them in place.

### Renaming without `moved` blocks

On Terraform before 1.8 and OpenTofu before 1.10, a `moved` block between two resource types fails with `Error: Resource type mismatch`. To rename on those versions, take each resource out of state under its old name and import it under the new one. Nothing is deleted in Hookdeck.

```sh
terraform state rm hookdeck_source.stripe
terraform import hookdeck_gateway_source.stripe src_xxx
```

Take the IDs from `terraform state show` before removing anything. A source auth is imported by the ID of its source.

The plan after the import shows an in-place update for each source and destination that sets `config`, because `config` is not read back on import. Applying it sends the same configuration again.

## `project_id` in State

Every source, source auth, destination, connection and transformation now has a `project_id` attribute, and it is always recorded in state. After the upgrade the provider fills it in from the project the resource already belongs to. The plan stays empty and no configuration change is needed. Terraform 1.0 and 1.1 list the new attribute under "Objects have changed outside of Terraform"; the plan still reports no changes.

The provider reads, updates and deletes a resource in the project recorded for it. If the API key is later swapped for a key of another project, the plan fails with `Project not reachable` and nothing is changed.

## `disabled_at` Is Read-Only

`disabled_at` on destinations and connections could be set in a v2 configuration, but the provider did not send it to Hookdeck, and an apply that set it failed with "Provider produced inconsistent result after apply". In v3 it is a read-only attribute, and setting it fails with `Invalid Configuration for Read-Only Attribute`. Remove it:

```terraform
resource "hookdeck_gateway_destination" "api" {
  name = "api"
  type = "HTTP"
  # disabled_at = "2025-01-01T00:00:00Z"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
  })
}
```

## Organization API Keys and Projects

v2 authenticates with a project API key. v3 accepts a project API key or an organization API key (prefix `hd_org_`), which can manage resources in several projects and create projects.

With a project API key nothing changes: every resource is in the key's project.

With an organization API key, either set `project_id` on the provider, so that everything is in that one project, or leave it unset and set `project_id` on every resource and data source. The [provider documentation](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs#projects) has an example of each setup, the rules, and how to create an organization API key.

### Moving an existing configuration to an organization API key

First upgrade to v3 with your project API key and run `terraform apply` once, so that every resource has its `project_id` in state. `terraform state show` on any resource prints it. Then switch the key in one of two ways. Both leave the plan empty.

Set the project on the provider:

```terraform
variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key    = var.hookdeck_org_api_key
  project_id = "tm_xxx" # the project_id in state
}

resource "hookdeck_gateway_source" "stripe" {
  name = "stripe"
  type = "STRIPE"
}
```

Or set it on every resource and data source:

```terraform
variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_org_api_key
}

resource "hookdeck_gateway_source" "stripe" {
  project_id = "tm_xxx" # the project_id in state
  name       = "stripe"
  type       = "STRIPE"
}
```

Setting a `project_id` different from the one in state replaces the resource.

## `hookdeck_gateway_project`

The new `hookdeck_gateway_project` resource creates, renames and deletes Event Gateway projects. It needs an organization API key with the `projects.write` scope. In 3.0 it manages the project's name; other project settings are left as they are. Deleting a project deletes everything in it.

```terraform
resource "hookdeck_gateway_project" "staging" {
  name = "staging"
}

resource "hookdeck_gateway_source" "stripe_staging" {
  project_id = hookdeck_gateway_project.staging.id
  name       = "stripe"
  type       = "STRIPE"
}
```

The data source of the same name looks up an existing Event Gateway project by `id` or by `name`, with either kind of key:

```terraform
data "hookdeck_gateway_project" "prod" {
  name = "prod"
}
```

## Import IDs

Resources are imported by ID. The ID can carry the project:

```sh
terraform import hookdeck_gateway_source.stripe src_xxx
terraform import hookdeck_gateway_source.stripe tm_xxx/src_xxx
```

With an organization API key and no provider `project_id`, the `<project_id>/<id>` form is required. In every other setup both forms work, and a project in the import ID has to be the provider's project.

Projects are imported by their ID:

```sh
terraform import hookdeck_gateway_project.prod tm_xxx
```

## Resources Deleted Outside Terraform

In v2, a resource deleted from the Hookdeck dashboard made `terraform plan` fail. In v3 it is removed from state and the plan proposes to create it again.

## API Version

The provider uses Hookdeck API version `2026-09-01` for every resource. In v2, destinations used `2026-09-01` and the other resources used `2025-07-01`. No attribute of a source, source auth, connection or transformation changed between the two.

## Troubleshooting

### `Error: Invalid resource type` ... `does not support resource type "hookdeck_gateway_source"`

The lock file still pins a v2 provider. Set the version constraint to `~> 3.0` and run `terraform init -upgrade`.

### `Error: Resource type mismatch`

Terraform before 1.8 and OpenTofu before 1.10 cannot move a resource to another type. Upgrade the CLI, keep the v2 names, or [rename without `moved` blocks](#renaming-without-moved-blocks).

### `Error: Invalid Configuration for Read-Only Attribute`

The configuration sets `disabled_at`. Remove it.

### `Error: Missing project_id`

The provider has an organization API key and no `project_id`, and a resource or data source does not set `project_id`. Set `project_id` on that resource or data source, or set `project_id` on the provider to put everything in one project.

The same error appears when destroying a `hookdeck_gateway_source_auth` whose state was last written by v2, with an organization API key and no provider `project_id`: that state does not record the project. Set `project_id` on the provider for that run. Applying once with v3 before switching keys avoids it, as described in [Moving an existing configuration to an organization API key](#moving-an-existing-configuration-to-an-organization-api-key).

### `Error: Project mismatch`

A resource or data source sets a `project_id` other than the provider's project. With a project API key or a provider `project_id`, everything is in that one project: remove `project_id` from the resource. To manage several projects with one provider configuration, use an organization API key without a provider `project_id`.

On a `hookdeck_gateway_source_auth`, the same error means its `project_id` differs from the project of its source.

### `Error: Project not reachable`

A resource is recorded in one project, and the provider now targets another project with a project API key. Nothing was changed. Do one of the following:

- Restore the API key or `project_id` that was changed by mistake.
- Stop managing the resource with `terraform state rm` or a `removed` block.
- Use an organization API key with access to both projects. The plan then replaces the resource in the new project.

### `Error: Project not accessible`, or `The API key cannot access project ...`

The Hookdeck API refused access to the project named in the message. The message starts with `The API key cannot access project` and says why: the project does not exist, it belongs to another organization, an organization API key has no access to it or lacks a scope, or a project API key belongs to another project.

The title is `Error: Project not accessible` when the provider checks the new project of a replacement while planning. When the API refuses a request for a resource, the same message appears under that request's title, such as `Error: Error reading source` or `Error: Error creating destination`.

Check `project_id` and the key. A resource that is already in state stays there.

### `Error: Missing project in import ID`

The provider has an organization API key and no `project_id`. Import as `<project_id>/<id>`.

### `Error: Organization API key required`

The `hookdeck_gateway_project` resource is used with a project API key. Use an organization API key, or read the project with the `hookdeck_gateway_project` data source.

### `Error: Unsupported project type`

The `hookdeck_gateway_project` resource or data source was given a project that is not an Event Gateway project.

### `Error: Unable to look up the API key's project`

With a project API key and no `project_id`, the provider asks the API which project the key belongs to, and that request failed. Check the API key. Setting `project_id` on the provider skips the lookup.
