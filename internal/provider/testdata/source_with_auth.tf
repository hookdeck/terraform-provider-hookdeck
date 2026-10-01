resource "hookdeck_gateway_source" "test" {
  name = "v3-src-%[1]s"
  type = "HTTP"
}

resource "hookdeck_gateway_source_auth" "test" {
  source_id = hookdeck_gateway_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "secret-%[1]s"
  })
}
