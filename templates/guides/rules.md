---
page_title: "Connection Rules"
description: "Connection Rules"
---

# Hookdeck Rules

A rule is a piece of instructional logic that dictates the behavior of events routed through a connection. There are 4 types of rules in Hookdeck:

## Retry rule

The retry rule determines the rate and limit of [automatic retries](https://hookdeck.com/docs/automatically-retry-events) on failed events.

| Retry Rule Element | Explanation                                                                                                                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Strategy           | A **linear** strategy means retries will occur at regular intervals; an **exponential** strategy means each retry will be delayed twice as long as the previous (1 hour, 2 hours, 4 hours...) |
| Interval           | The delay between each automatic retry                                                                                                                                                        |
| Count              | The number of automatic retries Hookdeck will attempt after an event fails                                                                                                                    |

> Automatic retries max out after one week, or {% $MAX_AUTOMATIC_RETRY_ATTEMPTS %} attempts – whichever comes first. Events can be [manually retried](https://hookdeck.com/docs/manually-retry-events) after exceeding this limit.

Here's what a connection with a linear retry strategy with five attempts per hour looks like:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      retry_rule = {
        count    = 5
        interval = 3600000
        strategy = "linear"
      }
    }
  ]
}
```

## Delay rule

The delay rule allows you to introduce a delay between the moment Hookdeck receives an event, and when it's forwarded to your destination.

Here's how to configure a connection with a 10-second delay:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      delay_rule = {
        delay = 10000
      }
    }
  ]
}
```

## Filter rule

The filter rule allows you to route webhooks based on the contents of their `Headers`, `Body`, `Query`, and/or `Path`.

For more information on how to set up filters, see our [filter documentation](https://hookdeck.com/docs/filters).

Here's how a connection with a filter look like:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      filter_rule = {
        body = {
          json = "{\"hello\":\"world\"}"
        }
        path = {
          string = "/api/webhook"
        }
        query = {
          boolean = false
        }
        headers = {
          number = 10
        }
      }
    }
  ]
}
```

As the Terraform provider expects a stringified JSON value for the JSON rule, there are some other approaches that you can use to configure your connection to your liking:

- Using Terraform's [`jsonencoded`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) for better readbility inline:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      filter_rule = {
        body = {
          json = jsonencode({
            hello = "world"
          })
        }
      }
    }
  ]
}
```

- Using Terraform's [`file`](https://developer.hashicorp.com/terraform/language/functions/file) to write your filter code in a separate file:

```hcl
resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      filter_rule = {
        body = {
          json = file("${path.module}/filters/my_connection_body.json")
        }
      }
    }
  ]
}
```

## Transformation rule

The transformation rule allows you to modify the payload of a webhook before it gets delivered to a destination.

For more information on how to set up transformations, see the [Hookdeck transformation documentation](https://hookdeck.com/docs/transformations).

To use transformation with Terraform, you must create a new transformation before using it with your connection. Here's an example transformation:

```hcl
resource "hookdeck_transformation" "my_transformation" {
  name = "my_transformation"
  code = "addHandler('transform', (request, context) => request);"
}

resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      transform_rule = {
        transformation_id = hookdeck_transformation.my_transformation.id
      }
    }
  ]
}
```

As the transformation code also expects a stringified function handler, you can also keep your transformation code in a separate file:

```hcl
resource "hookdeck_transformation" "my_transformation" {
  name = "my_transformation"
  code = file("${path.module}/transformations/my_transformation.js")
}
```

With the Hookdeck Terraform provider, you can keep your filter and transformation code in version control.
