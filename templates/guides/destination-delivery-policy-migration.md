---
page_title: "Migrating destination delivery policies"
description: "Breaking destination configuration changes for API version 2026-09-01"
---

# Destination delivery policy breaking change

Destination operations now use Hookdeck API version `2026-09-01`, matching the [API breaking change](https://hookdeck.com/docs/api#2026-09-01). This requires a **major provider release**. Existing Terraform configurations using the old fields must be edited before upgrading and applying. The provider does not translate legacy configuration or automatically rewrite state.

## Affected configuration

The JSON `config` on `hookdeck_destination` uses the new delivery policy structure for HTTP and Mock API destinations. Destinations without delivery rate or group settings need no configuration changes. Connections continue to reference destinations through `destination_id`; no connection configuration change is required.

| Previous config field | New config field |
| --- | --- |
| `rate_limit` | `delivery_policy.rate` |
| `rate_limit_period` | `delivery_policy.period` |
| `delivery_groups` | `delivery_policy.groups` |
| `delivery_groups.rate_limit` | `delivery_policy.groups.rate` |
| `delivery_groups.rate_limit_period` | `delivery_policy.groups.rate_period` |
| `delivery_groups.overrides.<value>.rate_limit` | `delivery_policy.groups.overrides.<value>.rate` |
| `delivery_groups.overrides.<value>.rate_limit_period` | `delivery_policy.groups.overrides.<value>.rate_period` |

Legacy fields, including fields set to `null`, fail configuration validation. Mixing old and new fields also fails. Group and override periods use `rate_period`; the destination-wide period uses `period`.

## Before

```hcl
resource "hookdeck_destination" "api" {
  name = "my-api"
  type = "HTTP"
  config = jsonencode({
    url               = "https://example.com/webhooks"
    rate_limit        = 100
    rate_limit_period = "minute"
    delivery_groups = {
      key               = "body.customer_id"
      rate_limit        = 5
      rate_limit_period = "second"
      overrides = {
        priority = {
          rate_limit        = 10
          rate_limit_period = "second"
        }
      }
    }
  })
}
```

## After

```hcl
resource "hookdeck_destination" "api" {
  name = "my-api"
  type = "HTTP"
  config = jsonencode({
    url = "https://example.com/webhooks"
    delivery_policy = {
      rate   = 100
      period = "minute"
      groups = {
        key         = "body.customer_id"
        rate        = 5
        rate_period = "second"
        overrides = {
          priority = {
            rate        = 10
            rate_period = "second"
          }
        }
      }
    }
  })
}
```

## Upgrade steps

1. Update each affected destination's JSON configuration using the mapping above. Preserve all existing rate, period, group, and override settings that should remain active.
2. Update your provider version constraint to the release containing this breaking change and run `terraform init -upgrade`.
3. Run `terraform plan` and review the destination configuration updates. Old field names produce a migration error, even if the old configuration has not changed.
4. Run `terraform apply`. The configuration update uses the existing destination ID; it does not require replacing the destination, removing state, or reimporting it. Terraform records your new configuration in state after applying it.

To disable all existing rate and group settings during migration, explicitly set `delivery_policy = null`. Simply deleting the old fields does not instruct the new API to clear their stored values. For a partial policy during migration, explicitly set unwanted `rate`, `period`, or `groups` members to `null`.

After migration, removing a previously configured policy member clears it on the API; removing `groups` disables groups, and removing an override replaces the group's overrides. Removing the entire `delivery_policy` also clears it.
