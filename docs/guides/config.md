---
page_title: "Source and destination config"
description: "How the provider compares the config of a source or destination with Hookdeck"
---

# Source and destination config

`config` on `hookdeck_gateway_source` and `hookdeck_gateway_destination` (and the deprecated `hookdeck_source` and `hookdeck_destination`) is a JSON object, written with `jsonencode`. From 3.1.0 the provider reads it back from Hookdeck, so a change made in the dashboard or through the API shows in the plan.

```terraform
resource "hookdeck_gateway_destination" "api" {
  name = "api"
  type = "HTTP"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
    delivery_policy = {
      rate   = 10
      period = "second"
    }
  })
}
```

## What is compared

Only the keys you set are compared with Hookdeck. In the example above those are `url` and, inside `delivery_policy`, `rate` and `period`.

- A key you set whose value differs in Hookdeck shows as a change, and the next apply restores your value.
- A key you do not set is not compared. This includes defaults Hookdeck adds and values set in the dashboard.
- Numbers and booleans written as strings (`"10"`, `"true"`) match the number or boolean Hookdeck returns.

## Seeing which key differs

`config` is sensitive, so the plan shows a change as `config = (sensitive value)`. To see the values, save the plan and print it as JSON:

```shell
terraform plan -out=tfplan
terraform show -json tfplan
```

The plan file and the JSON output include the secrets in `config`. Keep them out of tickets and CI logs.

## Removing a key

Removing a key from `config` clears it in Hookdeck on the next apply, or resets it to its default.

One exception: a list or true/false key that has no default keeps its value, for example `allowed_http_methods` on a `WEBHOOK` source. To change it, write the value you want.

To manage a key in the dashboard, leave it out of `config`. If the key was in `config` before, the apply that removes it clears it, so set it again in the dashboard after that apply. To keep every value as it is in Hookdeck, add `lifecycle { ignore_changes = [config] }`, which applies to all of `config`.

Leave `auth` out of a source's `config` when `hookdeck_gateway_source_auth` manages the source's authentication. With both, each resource changes the other's value back on every apply.

## Import

Importing a source or destination plans no change when `config` matches what Hookdeck has.

After an import, the provider does not know yet which keys are yours. Until an apply has changed `config`:

- Removing a key does not clear it in Hookdeck. To clear it, set it to `null`, apply, then remove it.
- Adding a key whose value Hookdeck already holds plans no change.

## Keys Hookdeck does not return

A top-level key that Hookdeck does not return for the type has no effect in Hookdeck: a misspelled key, or `url` on a `MOCK_API` destination. After the apply that sends it, plans ignore the key and warn on each run.

## Secrets in state

State holds the whole `config` Hookdeck returns, including secrets you did not write in your configuration, such as authentication tokens. The same goes for `config` on the source and destination data sources. The attribute is sensitive, and the state needs the protection you give any secret.

## Upgrading from 3.0

The first refresh stores Hookdeck's full configuration in state. The plan stays empty unless a key you set differs in Hookdeck. A value listed under [Known limitations](#known-limitations) plans a change on every run until you rewrite it as the table shows. Nothing else needs a configuration change.

After an apply that changes `config`, state holds your `config` until the next refresh replaces it with Hookdeck's full configuration. Terraform 1.0 and 1.1 list this under "Objects have changed outside of Terraform", and `-refresh-only` plans show it. Nothing changes in Hookdeck.

## Going back to 3.0.x

Once `terraform apply` (also with `-refresh-only`), `terraform refresh` or `terraform import` has run on 3.1.0, 3.0.x and earlier stop on sources and destinations with "Resource instance managed by newer provider version". Running only `terraform plan` on 3.1.0 leaves the state readable by 3.0.x.

3.1.0 does this on purpose: those versions treat every key in state as one you set, and their next apply would remove authentication, rate limits and other settings that are not in your `config`.

To go back, remove the sources and destinations from state with `terraform state rm` and import them again on 3.0.x. The first plan after the import shows `config` as a change. Applying it sends your `config` and clears nothing else in Hookdeck.

## Known limitations

A value that Hookdeck stores differently from how you wrote it shows as `config = (sensitive value)` on every plan, and applying does not make the plan empty. These configurations planned no change on 3.0.

| In `config` | What Hookdeck stores | What to write |
|---|---|---|
| `allowed_http_methods` on a source type whose methods Hookdeck fixes, such as `STRIPE` or `ZOOM` | its own list | Remove the key, or write the list Hookdeck returns, in the same order (`["POST"]` for `STRIPE` and `ZOOM`) |
| `custom_response` on a source type other than `WEBHOOK`, `HTTP` and `SENDGRID_INBOUND` | `null` | Remove the key |
| `path` on a `CLI` destination with surrounding spaces or without a leading `/` | the path trimmed, with a leading `/` | `"/path"` |
| `delivery_policy.period` with capitals or surrounding spaces | the value trimmed, in lower case | The value trimmed, in lower case |
| `url` on a destination with surrounding spaces | the URL trimmed | The URL without the spaces |
| A key set to `null` whose cleared value is a default: `delivery_timeout`, `auth`, `auth_type`, `delivery_policy` and `delivery_policy.period` on a destination | the default | Remove the key instead of setting it to `null` |
| A key inside `auth`, `custom_response` or `delivery_policy` that Hookdeck drops: a misspelled key, or a key of an authentication method other than the one in `auth_type` | nothing | Remove the key |
