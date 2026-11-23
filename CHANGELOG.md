## 3.1.0 (Unreleased)

FEATURES:

* New resource `hookdeck_outpost_project` creates an Outpost project and manages its configuration (`config` and `sensitive_config`), and new data source `hookdeck_outpost_project` reads a project and its whole configuration. See [Managing Outpost with Terraform](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/outpost). The resource needs an organization API key.
* New provider attribute `outpost_api_base` (environment variable `HOOKDECK_OUTPOST_API_BASE`) sets the Outpost API host. The default is `api.outpost.hookdeck.com`.
* `config` on sources and destinations is read from Hookdeck, so a change made outside Terraform to a key you set shows in the plan (see NOTES). Keys you do not set in `config`, including defaults Hookdeck adds, are kept in state and are not compared.
* The `hookdeck_gateway_source` and `hookdeck_gateway_destination` data sources (and the deprecated `hookdeck_source` and `hookdeck_destination`) return `config`.
* Importing a source or destination plans no change when `config` matches what Hookdeck has.

NOTES:

* Behaviour change: a change made in the dashboard or API to a key you set in `config` now shows in the plan as `config = (sensitive value)`, and the next apply restores the configured value. To keep a value managed outside Terraform, remove its key from `config`, or add `lifecycle { ignore_changes = [config] }`. To see which key differs, run `terraform plan -out=tfplan` and read `terraform show -json tfplan`, which prints the values.
* The first refresh after upgrading stores Hookdeck's full configuration in state. The plan stays empty unless a key you set in `config` differs in Hookdeck. Upgrading needs no configuration change.
* Going back to 3.0.x: once `terraform apply` (also with `-refresh-only`) or `terraform refresh` has run on 3.1.0, 3.0.x and earlier stop on its sources and destinations with "Resource instance managed by newer provider version". This is deliberate. Those versions would treat every key stored in state, including the ones Hookdeck returned, as one you had set, and their next apply would remove authentication, rate limits and other settings that are not in your `config`. To go back, remove the sources and destinations from state with `terraform state rm` and import them again on 3.0.x. Running only `terraform plan` on 3.1.0 leaves the state readable by 3.0.x.
* `config` in state and in the data sources now holds the secrets Hookdeck returns, such as authentication tokens. It stays sensitive.
* After an apply that changes `config`, state holds exactly your `config`, and the next refresh replaces it with Hookdeck's full configuration. Terraform 1.0 and 1.1 list this under "Objects have changed outside of Terraform" on the next plan, and `-refresh-only` plans show it. The plan still reports no changes.
* A key in `config` that Hookdeck does not return for the type (a typo, a key of another type, or `url` on a `MOCK_API` destination) has no effect in Hookdeck. After the apply that sends it, plans ignore it and warn on each run.
* Numbers and booleans written as strings (`"10"`, `"true"`) match the number or boolean Hookdeck returns.
* After an import, removing a key from `config` does not clear it in Hookdeck until an apply has changed `config`; set the key to `null` to clear it before then (see the limitation on `null` below). Adding a key whose value Hookdeck already holds plans no change until then either.
* Known limitations: a value in `config` that Hookdeck stores differently from how you wrote it shows as `config = (sensitive value)` on every plan, and applying does not make the plan empty. These configurations planned no change on 3.0. The known cases:
  * `allowed_http_methods` on a source type whose methods Hookdeck fixes, such as `STRIPE` or `ZOOM`, or that has no method setting, such as `ZENDESK`: Hookdeck stores its own list or `null`. Remove the key, or for a fixed list write exactly that list.
  * `custom_response` on a source type that does not support it (any type other than `WEBHOOK`, `HTTP` and `SENDGRID_INBOUND`): Hookdeck stores `null`. Remove the key.
  * `path` on a `CLI` destination with surrounding spaces or without a leading `/`: Hookdeck trims it and adds the `/`. Write `"/path"`.
  * `delivery_policy.period` with capitals or surrounding spaces: Hookdeck stores it trimmed and in lower case. Write it that way.
  * `url` on a destination with surrounding spaces: Hookdeck trims them. Remove the spaces.
  * A key set to `null` whose cleared value in Hookdeck is a default (`delivery_timeout`, `auth`, `auth_type`, `delivery_policy` and `delivery_policy.period` on a destination): Hookdeck returns the default. Remove the key once the apply has cleared it.

BUG FIXES:

* Destroying a source, source auth, destination, connection or transformation that no longer exists in Hookdeck succeeds instead of failing, for example with `terraform destroy -refresh=false`.
* Removing `rules` from a connection's configuration clears the rules in Hookdeck. Before, the apply failed with "Provider produced inconsistent result after apply". Rules added outside Terraform to a connection configured without `rules` are now cleared the same way.
* `hookdeck_gateway_project` is no longer removed from state, and created again on the next apply, when the API key cannot reach the project, for example a key from another organization. Terraform reports that the key cannot access the project and keeps it in state. Renaming or destroying such a project fails with the same error; before, `terraform destroy -refresh=false` removed it from state while it still existed. A project deleted outside Terraform is still removed from state, and destroying it succeeds. With an organization API key, importing such a project, or reading it by `id` with the `hookdeck_gateway_project` data source, gives the same error instead of "not found".

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