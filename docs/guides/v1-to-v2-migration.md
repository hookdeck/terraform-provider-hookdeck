---
page_title: "Migrating from v1.x to v2.x"
description: "How to migrate from v1.x to v2.x of the Hookdeck Terraform Provider"
---

The Hookdeck API version `2025-07-01` introduces a [breaking change](https://hookdeck.com/docs/api#2025-07-01), [Transform](https://hookdeck.com/docs/transformations) and [Filter](https://hookdeck.com/docs/filters) rules in a connection are now **executed in the order defined**. Previously, the execution order was fixed: **Transform rules always ran before Filter rules**.

Hookdeck has automatically migrated all existing rules to match the previous behavior (Transform before Filter). However, your **Terraform configuration may now show diffs** if it defines rules in a different order than what is stored remotely.

This guide walks through how to adjust your configuration to match the new behavior.

## Summary of Change

- **What changed**: The `rules` list in a `hookdeck_connection` is now **ordered**. `transform_rule` and `filter_rule` blocks are executed in the order they appear.
- **Which rules are affected**: Only `transform_rule` and `filter_rule` blocks respect the order. Other rule blocks (e.g., `retry_rule`, `delay_rule`) are not affected as their execution order is not important.
- **Default behavior**: Existing Hookdeck connections have been migrated so that `transform_rule` blocks appear before `filter_rule` blocks in the `rules` list—preserving legacy behavior.
- **Terraform implications**: If your configuration has a `filter_rule` before a `transform_rule`, Terraform will detect this as a change.

## What You Should Do

Update your Terraform `rules` definition to ensure `transform_rule` blocks come **before** `filter_rule` blocks, unless you intentionally want the new ordering behavior.

### Example

The `v1.x` provider introduced the `rules` attribute with the modern syntax for `filter_rule` and `transform_rule`. The breaking change in `v2.x` is that the *order* of these rules now matters.

For example, if your configuration defined a `filter_rule` before a `transform_rule`:

```hcl
resource "hookdeck_transformation" "my_transformation" {
  name = "my-transformation"
  code = file("${path.module}/transformation.js")
}

resource "hookdeck_connection" "my_connection" {
  # ... other attributes like source_id and destination_id

  rules = [
    {
      // This filter rule runs first in the list...
      filter_rule = {
        body = {
          // Using jsonencode is recommended for readability
          json = jsonencode({
            status = "active"
          })
        }
      }
    },
    {
      // ...followed by this transform rule.
      transform_rule = {
        transformation_id = hookdeck_transformation.my_transformation.id
      }
    }
  ]
}
```

Because Hookdeck automatically migrated existing connections to have `transform` rules run before `filter` rules, the above configuration will produce a `terraform plan` diff.

To align your configuration with the migrated state and avoid the diff, you must reorder the rules in your file:

```hcl
resource "hookdeck_transformation" "my_transformation" {
  name = "my-transformation"
  code = file("${path.module}/transformation.js")
}

resource "hookdeck_connection" "my_connection" {
  # ... other attributes like source_id and destination_id

  rules = [
    {
      // The transform rule is now first, matching the legacy behavior.
      transform_rule = {
        transformation_id = hookdeck_transformation.my_transformation.id
      }
    },
    {
      // The filter rule is now second.
      filter_rule = {
        body = {
          json = jsonencode({
            status = "active"
          })
        }
      }
    }
  ]
}
```

This will preserve the same behavior as before the API update and resolve any `terraform plan` diffs.

## Optional: Leverage the New Behavior

This update also unlocks new use cases:

- Run a **filter before a transformation** to skip unnecessary processing.
- Use a transformation **after** a filter to clean up headers or payloads before delivery.

If you're intentionally using the new ordering capability, no migration is needed—just ensure your rule order in Terraform matches your intended logic.

## Final Notes

- The Hookdeck provider does **not** modify rule order—your configuration must match your intended logic.
- If you see diffs after upgrading, they are likely due to mismatched rule order and can be resolved by reordering locally.
- The provider will output a warning message if it detects that your existing rule definitions have a filter rule before a transformation rule.

For questions or help migrating, please [raise an issue on the Hookdeck Terraform Provider GitHub repo](https://github.com/hookdeck/terraform-provider-hookdeck/issues).
