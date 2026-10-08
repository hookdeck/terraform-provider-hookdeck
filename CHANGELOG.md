## 3.0.0 (Unreleased)

See the [v2 to v3 migration guide](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/v2-to-v3-migration).

BREAKING CHANGES:

* Event Gateway resources and data sources are renamed with a `gateway_` prefix: `hookdeck_gateway_source`, `hookdeck_gateway_source_auth`, `hookdeck_gateway_destination`, `hookdeck_gateway_connection`, `hookdeck_gateway_transformation`. The v2 names remain as deprecated aliases until v4. Rename with `moved` blocks; no resources are recreated. `moved` blocks between resource types need Terraform 1.8 or OpenTofu 1.10, or later; the guide has a route for earlier versions.
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