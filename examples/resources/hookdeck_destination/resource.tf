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
