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
  name = "example"
  type = "HTTP"
  config = jsonencode({
    url       = "https://example.test/webhook"
    auth_type = "BASIC_AUTH"
    auth = {
      username = "username"
      password = "password"
    }
    rate_limit        = 10
    rate_limit_period = "concurrent"
  })
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A unique, human-friendly name for the destination

### Optional

- `config` (String, Sensitive) Destination configuration
- `description` (String) Description for the destination
- `disabled_at` (String) Date the destination was disabled
- `type` (String) Type of the destination

### Read-Only

- `created_at` (String) Date the destination was created
- `id` (String) ID of the destination
- `team_id` (String) ID of the workspace
- `updated_at` (String) Date the destination was last updated

## Import

Import is supported using the following syntax:

```shell
$ terraform import hookdeck_destination.example <destination_id>
```
