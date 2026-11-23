## 3.1.0 (Unreleased)

NOTES:

* Sources and destinations: `config` is now read from Hookdeck. A key you set that was changed in the dashboard or API shows in the plan as `config = (sensitive value)`, and the next apply restores your value. Keys you do not set are not compared. See [Source and destination config](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/config).
* Sources and destinations: removing a key from `config` clears it in Hookdeck or resets it to its default (the guide lists the exceptions). To keep values managed outside Terraform, add `lifecycle { ignore_changes = [config] }`.
* Sources and destinations: a few values that Hookdeck stores differently from how they were written plan a change on every run until rewritten. The guide lists them. Other configurations need no change.
* Sources and destinations: once `terraform apply`, `terraform refresh` or `terraform import` has run on 3.1.0, 3.0.x and earlier stop with "Resource instance managed by newer provider version". The guide describes how to go back.
* `config` in state, and in the source and destination data sources, now holds the secrets Hookdeck returns, such as authentication tokens. It stays sensitive.
* Connections: rules added outside Terraform to a connection configured without `rules` show in the plan, and the next apply removes them. To keep them, add `lifecycle { ignore_changes = [rules] }`.

FEATURES:

* New resource and data source `hookdeck_outpost_project`: create an Outpost project and manage its configuration. The resource needs an organization API key. See [Managing Outpost with Terraform](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/outpost). New provider attribute `outpost_api_base` (`HOOKDECK_OUTPOST_API_BASE`) sets the Outpost API host. It does not follow `api_base`: set both for a non-default host.
* The `hookdeck_gateway_source` and `hookdeck_gateway_destination` data sources (and the deprecated `hookdeck_source` and `hookdeck_destination`) return `config`.
* Importing a source or destination plans no change when `config` matches Hookdeck.

BUG FIXES:

* Destroying a source, source auth, destination, connection or transformation that no longer exists in Hookdeck succeeds. So does destroying one whose project was deleted; with an organization API key, a refresh removes it from state.
* API reads stop when Terraform is interrupted or a resource's `timeouts` pass. A write already sent to Hookdeck still finishes, so the result reaches state.
* Removing `rules` from a connection clears its rules in Hookdeck. Before, the apply failed with "Provider produced inconsistent result after apply".
* Rules removed from a connection outside Terraform show in the plan, and the next apply restores them.
* `hookdeck_gateway_project` stays in state when the API key cannot reach the project, for example a key from another organization, instead of being created again on the next apply. Plan, rename and destroy fail with an error saying the key cannot access the project. Importing such a project, or reading it by `id` with the data source, gives the same error instead of "not found".

## 3.0.0 (2026-11-17)

See the [v2 to v3 migration guide](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/v2-to-v3-migration).

BREAKING CHANGES:

* Event Gateway resources and data sources are renamed with a `gateway_` prefix: `hookdeck_gateway_source`, `hookdeck_gateway_source_auth`, `hookdeck_gateway_destination`, `hookdeck_gateway_connection`, `hookdeck_gateway_transformation`, and data sources `hookdeck_gateway_source`, `hookdeck_gateway_destination`, `hookdeck_gateway_connection`. The v2 names remain as deprecated aliases until v4. Rename with `moved` blocks; no resources are recreated. `moved` blocks between resource types need Terraform 1.8 or OpenTofu 1.10, or later; the guide has a route for earlier versions.
* The new `gateway_` names have no `team_id` attribute; use `project_id`, which holds the same value. The v2 names keep `team_id`.
* `disabled_at` on `hookdeck_gateway_destination` and `hookdeck_gateway_connection` is read-only. v2 accepted it in configuration but did not send it to Hookdeck. Remove it from configuration if set.

FEATURES:

* Organization API keys (`hd_org_` prefix), which reach several projects.
* Source, source auth, destination, connection and transformation resources, and their data sources, have a `project_id` attribute, always recorded in state. The provider also accepts `project_id` (or `HOOKDECK_PROJECT_ID`).
  * With a project API key or a provider `project_id`, every resource is in that project. A different `project_id` on a resource is an error when planning.
  * With an organization API key and no provider `project_id`, every resource and data source sets `project_id`.
  * Changing a resource's project replaces the resource.
  * A resource is read, updated and deleted in the project recorded in state. A key that cannot reach that project fails the plan.
  * State written by v2 needs no change: `project_id` is filled in and the plan is empty.
* New resource `hookdeck_gateway_project` (name only in this release) and data source `hookdeck_gateway_project` (lookup by `id` or `name`), for Event Gateway projects. The resource needs an organization API key.
* Import accepts `<project_id>/<id>` in addition to `<id>`. `hookdeck_gateway_source_auth` can be imported by the ID of its source.

BUG FIXES:

* A resource deleted outside Terraform is removed from state and recreated on the next apply instead of failing the plan.
* The provider address reported to Terraform is `registry.terraform.io/hookdeck/hookdeck`.
* The API key is no longer written to debug logs.

NOTES:

* Every resource uses Hookdeck API version `2026-09-01`. In v2 only destinations did.

See https://github.com/hookdeck/terraform-provider-hookdeck/releases