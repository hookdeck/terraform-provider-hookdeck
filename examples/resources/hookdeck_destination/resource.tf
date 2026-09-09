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
    delivery_policy = {
      rate   = 10
      period = "concurrent"
      groups = {
        key         = "body.customer_id"
        rate        = 5
        rate_period = "second"
      }
    }
  })
}
