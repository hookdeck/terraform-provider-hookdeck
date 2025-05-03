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
