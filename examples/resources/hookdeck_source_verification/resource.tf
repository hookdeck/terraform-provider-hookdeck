resource "hookdeck_source_verification" "verification_stripe" {
  source_id = hookdeck_source.source_example.id
  verification = {
    stripe = {
      webhook_secret_key = jsondecode(webhook_registration.webhook_stripe.register.response).body.secret
    }
  }
}
