---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hookdeck_destination Resource - terraform-provider-hookdeck"
subcategory: ""
description: |-
  Destination Resource
---

# hookdeck_destination (Resource)

Destination Resource

## Example Usage

```terraform
resource "hookdeck_destination" "example" {
  name        = "example"
  description = "example destination"
  url         = "https://mock.hookdeck.com"
  http_method = "POST"
  rate_limit = {
    limit  = 10
    period = "second"
  }
  auth_method = {
    api_key = {
      api_key = var.destination_api_key
      key     = "x-webhook-key"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A unique, human-friendly name for the destination

### Optional

- `auth_method` (Attributes) Config for the destination's auth method (see [below for nested schema](#nestedatt--auth_method))
- `cli_path` (String) Path for the CLI destination
- `description` (String) Description for the destination
- `http_method` (String) must be one of ["GET", "POST", "PUT", "PATCH", "DELETE"]
HTTP method used on requests sent to the destination, overrides the method used on requests sent to the source.
- `path_forwarding_disabled` (Boolean)
- `rate_limit` (Attributes) Rate limit (see [below for nested schema](#nestedatt--rate_limit))
- `url` (String) HTTP endpoint of the destination

### Read-Only

- `archived_at` (String) Date the destination was archived
- `created_at` (String) Date the destination was created
- `id` (String) ID of the destination
- `team_id` (String) ID of the workspace
- `updated_at` (String) Date the destination was last updated

<a id="nestedatt--auth_method"></a>
### Nested Schema for `auth_method`

Optional:

- `api_key` (Attributes) API Key (see [below for nested schema](#nestedatt--auth_method--api_key))
- `basic_auth` (Attributes) Basic Auth (see [below for nested schema](#nestedatt--auth_method--basic_auth))
- `bearer_token` (Attributes) Bearer Token (see [below for nested schema](#nestedatt--auth_method--bearer_token))
- `custom_signature` (Attributes) Custom Signature (see [below for nested schema](#nestedatt--auth_method--custom_signature))
- `hookdeck_signature` (Attributes) Hookdeck Signature (see [below for nested schema](#nestedatt--auth_method--hookdeck_signature))

<a id="nestedatt--auth_method--api_key"></a>
### Nested Schema for `auth_method.api_key`

Required:

- `api_key` (String, Sensitive) API key for the API key auth
- `key` (String) Key for the API key auth

Optional:

- `to` (String) must be one of ["header", "query"]
Whether the API key should be sent as a header or a query parameter


<a id="nestedatt--auth_method--basic_auth"></a>
### Nested Schema for `auth_method.basic_auth`

Required:

- `password` (String, Sensitive) Password for basic auth
- `username` (String) Username for basic auth


<a id="nestedatt--auth_method--bearer_token"></a>
### Nested Schema for `auth_method.bearer_token`

Required:

- `token` (String, Sensitive) Token for the bearer token auth


<a id="nestedatt--auth_method--custom_signature"></a>
### Nested Schema for `auth_method.custom_signature`

Required:

- `key` (String) Key for the custom signature auth

Optional:

- `signing_secret` (String, Sensitive) Signing secret for the custom signature auth. If left empty a secret will be generated for you.


<a id="nestedatt--auth_method--hookdeck_signature"></a>
### Nested Schema for `auth_method.hookdeck_signature`



<a id="nestedatt--rate_limit"></a>
### Nested Schema for `rate_limit`

Required:

- `limit` (Number) Limit event attempts to receive per period. Max value is workspace plan's max attempts thoughput.
- `period` (String) must be one of ["second", "minute", "hour"]
Period to rate limit attempts

## Import

Import is supported using the following syntax:

```shell
$ terraform import hookdeck_destination.example <destination_id>
```
