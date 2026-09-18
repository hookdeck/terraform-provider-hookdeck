## 3.0.0 (Unreleased)

See the [v2 to v3 migration guide](docs/guides/v2-to-v3-migration.md).

BREAKING CHANGES:

* Event Gateway resources and data sources are renamed with a `gateway_` prefix: `hookdeck_gateway_source`, `hookdeck_gateway_source_auth`, `hookdeck_gateway_destination`, `hookdeck_gateway_connection`, `hookdeck_gateway_transformation`. The v2 names remain as deprecated aliases until v4. Rename with `moved` blocks; no resources are recreated. Requires Terraform 1.8 or later.
* `disabled_at` on `hookdeck_gateway_destination` and `hookdeck_gateway_connection` is read-only. It was accepted in v2 configuration but never sent.
* Every resource targets Hookdeck API version `2026-09-01`.

FEATURES:

* Organization API keys (`hd_org_` prefix). Every project-scoped resource and data source accepts `project_id`; the provider accepts a default `project_id` (or `HOOKDECK_PROJECT_ID`). Precedence: resource, provider, then the key's own project.
* New resource `hookdeck_gateway_project` and data source `hookdeck_gateway_project` (lookup by `id` or `name`).
* Import accepts `<project_id>/<id>` in addition to `<id>`.

BUG FIXES:

* A resource deleted outside Terraform is removed from state and recreated on the next apply instead of failing the plan.
* The provider address reported to Terraform is `registry.terraform.io/hookdeck/hookdeck`.
* The API key is no longer written to debug logs.

See https://github.com/hookdeck/terraform-provider-hookdeck/releases