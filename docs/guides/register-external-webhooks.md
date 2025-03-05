---
page_title: "Register External Webhooks"
description: "Register External Webhooks with Hookdeck Provider"
---

# Register External Webhooks

The Hookdeck Terraform provider supports registering and unregistering webhooks with your external service providers, such as Stripe, Shopify, etc. It's service-agnostic, so if there's an endpoint to manage webhooks programmatically, you should be able to configure it as part of your workflow.

## Setup

For example, let's say you're using Stripe for payment and want to listen to `charge.succeeded` and `charge.failed` webhook events.

Let's start with a Hookdeck connection to start listening to incoming webhooks from Stripe:

```hcl
resource "hookdeck_source" "stripe" {
  name = "stripe"
}

resource "hookdeck_destination" "payment_service" {
  name = "my_destination"
  url  = "https://api.my-app.com/webhooks/stripe"
}

resource "hookdeck_connection" "stripe_payment_service" {
  source_id      = hookdeck_source.stripe.id
  destination_id = hookdeck_destination.payment_service.id
}
```

## Register Stripe webhook

Stripe provides API endpoints to [create a new webhook](https://stripe.com/docs/api/webhook_endpoints/create) and [delete an existing webhook](https://stripe.com/docs/api/webhook_endpoints/delete), which we will use in this example.

```hcl
resource "hookdeck_webhook_registration" "stripe_webhook_registration" {
  provider = hookdeck

  register = {
    request = {
      method = "POST"
      url    = "https://api.stripe.com/v1/webhook_endpoints"
      headers = jsonencode({
        "content-type" = "application/json"
        authorization  = "Bearer <STRIPE_SECRET_KEY>"
      })
      body = jsonencode({
        url = hookdeck_source.stripe.url
        enabled_events = [
          "charge.failed",
          "charge.succeeded"
        ]
      })
    }
  }
  unregister = {
    request = {
      method = "DELETE"
      url    = "https://api.stripe.com/v1/webhook_endpoints/{{.register.response.body.id}}"
      headers = jsonencode({
        authorization  = "Bearer <STRIPE_SECRET_KEY>"
      })
    }
  }
}
```

For many APIs, you will need the ID of the registered webhook to unregister. You can usually get the ID of the newly registered webhook from the register API response itself. Because of that, the Hookdeck Terraform provider will save that response for use in the `unregister` flow. When configuring your `unregister` property, the string values are string templates where you'll have access to `.register.response` data, which is the HTTP response from the original registration API request.

## Use webhook secret to verify with Hookdeck

Another way you can use the `hookdeck_webhook_registration` resource is to configure Hookdeck [source verification](https://hookdeck.com/docs/signature-verification) as part of your Terraform workflow. With the `hookdeck_webhook_registration` resource above, you can now configure Hookdeck verification like so:

```hcl
resource "hookdeck_source_auth" "stripe_source_auth" {
  source_id = hookdeck_source.stripe.id
  auth = jsonencode({
    webhook_secret_key = jsondecode(hookdeck_webhook_registration.stripe_webhook_registration.register.response).body.secret
  })
}
```

As mentioned in the section earlier, the provider will save the response from the registration request to be used later (unregister flow). For some APIs, that response will also contain secret information you can use for verification purposes. As the `webhook_registration` is provider-agnostic, it saves the response in a stringified JSON with two fields, "body" and "headers". When using that data in the unregister flow, the provider constructs that response and uses it in the template. When using the response in other resources, you will need to decode the stringified JSON using [`jsondecode`](https://developer.hashicorp.com/terraform/language/functions/jsondecode) like the example above.

Putting everything together to register Stripe webhook with Hookdeck source with source verification, here's how your Terraform code will look:

```hcl
# Configure Hookdeck source, destination, and connection

resource "hookdeck_source" "stripe" {
  name = "stripe"
}

resource "hookdeck_destination" "payment_service" {
  name = "my_destination"
  url  = "https://api.my-app.com/webhooks/stripe"
}

resource "hookdeck_connection" "stripe_payment_service" {
  source_id      = hookdeck_source.stripe.id
  destination_id = hookdeck_destination.payment_service.id
}

# Register Stripe webhook

resource "hookdeck_webhook_registration" "stripe" {
  provider = hookdeck

  register = {
    request = {
      method = "POST"
      url    = "https://api.stripe.com/v1/webhook_endpoints"
      headers = jsonencode({
        "content-type" = "application/json"
        authorization  = "Bearer <STRIPE_SECRET_KEY>"
      })
      body = jsonencode({
        url = hookdeck_source.stripe.url
        enabled_events = [
          "charge.failed",
          "charge.succeeded"
        ]
      })
    }
  }
  unregister = {
    request = {
      method = "DELETE"
      url    = "https://api.stripe.com/v1/webhook_endpoints/{{.register.response.body.id}}"
      headers = jsonencode({
        authorization  = "Bearer <STRIPE_SECRET_KEY>"
      })
    }
  }
}

# Configure source verification

resource "hookdeck_source_verification" "stripe_verification" {
  source_id = hookdeck_source.stripe.id
  verification = {
    stripe = {
      webhook_secret_key = jsondecode(hookdeck_webhook_registration.stripe.register.response).body.secret
    }
  }
}
```
