resource "hookdeck_webhook_registration" "webhook_stripe" {
  register = {
    request = {
      method = "POST"
      url    = "https://api.stripe.com/v1/webhook_endpoints"
      headers = jsonencode({
        "content-type" = "application/json"
        authorization  = "Bearer ${var.stripe_secret_key}"
      })
      body = jsonencode({
        url            = hookdeck_source.source_example.url
        enabled_events = ["charge.failed", "charge.succeeded"]
      })
    }
  }
  unregister = {
    request = {
      method = "DELETE"
      url    = "https://api.stripe.com/v1/webhook_endpoints/{{.register.response.body.id}}"
      headers = jsonencode({
        authorization = "Bearer ${var.stripe_secret_key}"
      })
    }
  }
}
