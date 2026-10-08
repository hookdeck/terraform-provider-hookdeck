resource "hookdeck_gateway_destination" "test" {
  name        = "v3-dst-%[1]s"
  type        = "HTTP"
  disabled_at = "2025-01-01T00:00:00Z"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}
