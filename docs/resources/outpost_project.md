---
page_title: "hookdeck_outpost_project Resource - terraform-provider-hookdeck"
subcategory: ""
description: |-
  Outpost project. Managing a project requires an organization API key with projects.write. Creating a project waits until it is ready, then applies config and sensitive_config. Deleting a project deletes everything in it.
---

# hookdeck_outpost_project (Resource)

Outpost project. Managing a project requires an organization API key with `projects.write`. Creating a project waits until it is ready, then applies `config` and `sensitive_config`. Deleting a project deletes everything in it.

See [Managing Outpost with Terraform](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/outpost) for a walk-through.

## Example Usage

```terraform
variable "otlp_headers" {
  type      = string
  sensitive = true
}

resource "hookdeck_outpost_project" "example" {
  name = "production-outpost"

  # Outpost environment variables to manage. Keys you leave out keep their value.
  config = {
    TOPICS                      = "order.created,order.updated"
    RETRY_INTERVAL_SECONDS      = "60"
    PORTAL_ORGANIZATION_NAME    = "Acme"
    OTEL_SERVICE_NAME           = "acme-outpost"
    OTEL_EXPORTER_OTLP_ENDPOINT = "https://otel.example.com"
  }

  # Values that are secrets or can carry credentials.
  sensitive_config = {
    OTEL_EXPORTER_OTLP_HEADERS = var.otlp_headers
  }

  # Deleting a project deletes everything in it.
  # lifecycle {
  #   prevent_destroy = true
  # }
}
```

## Configuration keys

The keys of `config` and `sensitive_config` are Outpost environment variable names. The [configuration reference](https://hookdeck.com/docs/outpost/self-hosting/configuration) lists them with their values. The reference also covers self-hosted deployments: some of the settings it lists are managed by Hookdeck, and Hookdeck rejects them when they are set on a project.

- Put a key in `config` or in `sensitive_config`, not in both. Keys whose values are secrets or can carry credentials must go in `sensitive_config`. These keys are `DESTINATIONS_PROXY_URL`, `OTEL_EXPORTER_OTLP_HEADERS`, `PUBLISH_AWS_SQS_SECRET_ACCESS_KEY`, `PUBLISH_AZURE_SERVICEBUS_CONNECTION_STRING`, `PUBLISH_GCP_PUBSUB_SERVICE_ACCOUNT_CREDENTIALS` and `PUBLISH_RABBITMQ_SERVER_URL`. A URL that includes a user or password also goes in `sensitive_config`, whatever its key. `terraform plan` fails if `config` holds one of these keys or such a URL. Values in `sensitive_config` are hidden in plan output and stored in the Terraform state.
- The provider does not check values. Hookdeck rejects an invalid value at apply and the message names the key. A key the provider does not know gets a warning at plan, which usually means a typo. The warning does not stop the plan, so a key added to Outpost after this provider release still works; Hookdeck rejects a key that does not exist when you apply.
- `OPERATOR_EVENTS_HTTP_URL` and `OPERATOR_EVENTS_TOPICS` are set by operator event destinations, and `DESTINATIONS_WEBHOOK_PROXY_URL` is a deprecated alias of `DESTINATIONS_PROXY_URL`. They cannot be set here.
- Only the keys you set are managed. Other keys keep their value, and changes to them outside Terraform do not show in the plan; a change to a managed key does. Removing a key from the configuration returns it to its default. The [`hookdeck_outpost_project` data source](https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/data-sources/outpost_project) returns the whole configuration, defaults included.

## Portal custom domain

Set `portal_custom_domain` to serve the project's portal on your own hostname.

```terraform
resource "hookdeck_outpost_project" "example" {
  name                 = "production-outpost"
  portal_custom_domain = "portal.example.com"
}

output "portal_dns_records" {
  value = hookdeck_outpost_project.example.portal_custom_domain_status.verification_records
}
```

- Applying adds the domain after the project is configured. `portal_custom_domain_status.verification_records` then lists the DNS records to create at your DNS provider, each with a `type`, `name` and `value`. These records verify the hostname and issue its certificate. `verification_records` is unknown until the domain is added, so if you manage your DNS in Terraform, apply the project first (for example with `-target`), then the records.
- `portal_custom_domain_status.status` is `pending` until the hostname is verified, then `active`, which means the hostname is verified. `ssl_status` is the state of the domain's certificate: `pending_validation` until the records are in place, `active` once it is issued. `verification_errors` lists the reasons the API reports when verification fails. Terraform refreshes the status on every plan.
- Changing the hostname deletes the current domain and adds the new one, which needs new DNS records. Removing `portal_custom_domain` deletes the domain.
- A domain added or changed outside Terraform shows in the plan and is removed or replaced on the next apply, unless `portal_custom_domain` is set to it.
- A hostname serves one project at a time: adding one that another project uses fails.
- Deleting the project deletes its domain first, which frees the hostname.

## Behavior

- Creating a project waits until it is ready, then applies the configuration. If provisioning is slow, raise `timeouts.create` (20 minutes by default).
- The project restarts to apply a configuration change and may be briefly unavailable. An apply that changes no configuration value, such as a rename, does not restart it.
- If a create fails after the project exists, for example because Hookdeck rejected a configuration value, the project stays in state as tainted and the next apply replaces it.
- Deleting the project deletes its tenants, destinations and events.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the project

### Optional

- `config` (Map of String) Outpost configuration to manage, keyed by Outpost environment variable name (for example `TOPICS` or `RETRY_INTERVAL_SECONDS`). Only the keys set here are managed and tracked; other keys keep their value. Removing a key restores its default. Keys that can carry credentials belong in `sensitive_config`.
- `portal_custom_domain` (String) Hostname that serves the project's portal, for example `portal.example.com`. Changing it deletes the current domain and adds the new one. `portal_custom_domain_status` lists the DNS records to create. A domain added outside Terraform is removed on the next apply unless it is set here.
- `sensitive_config` (Map of String, Sensitive) Outpost configuration to manage like `config`, for keys whose values are secrets or can carry credentials, such as `OTEL_EXPORTER_OTLP_HEADERS`, `DESTINATIONS_PROXY_URL` and the `PUBLISH_*` connection strings and secret keys. Values are stored in the Terraform state.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) Date the project was created
- `id` (String) ID of the project
- `organization_id` (String) ID of the organization the project belongs to
- `portal_custom_domain_status` (Attributes) State of the portal custom domain. Null when no domain is set. Refreshed on every plan, so it moves from `pending` to `active` once the DNS records are in place. (see [below for nested schema](#nestedatt--portal_custom_domain_status))
- `type` (String) Project type, always `outpost`
- `updated_at` (String) Date the project was last updated

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--portal_custom_domain_status"></a>
### Nested Schema for `portal_custom_domain_status`

Read-Only:

- `ssl_status` (String) Certificate status, for example `pending_validation` or `active`
- `status` (String) Domain status: `pending` until the hostname is verified, then `active`
- `verification_errors` (List of String) Reasons the domain could not be verified, if any
- `verification_records` (Attributes List) DNS records to create to verify the domain and issue its certificate (see [below for nested schema](#nestedatt--portal_custom_domain_status--verification_records))

<a id="nestedatt--portal_custom_domain_status--verification_records"></a>
### Nested Schema for `portal_custom_domain_status.verification_records`

Read-Only:

- `name` (String) Record name
- `type` (String) Record type, for example `txt`
- `value` (String) Record value

## Import

Import is supported using the following syntax:

```shell
$ terraform import hookdeck_outpost_project.example <project_id>
```

Import fills the project's name, metadata and portal custom domain. `config` and `sensitive_config` stay empty, so the first plan lists the keys in your configuration as additions. Applying sends nothing when their values already match the project.
