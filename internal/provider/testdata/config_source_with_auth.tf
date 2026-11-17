resource "hookdeck_gateway_source" "test" {
  name = "tf-acc-cfg-src-%[1]s"
  type = "HTTP"
  config = jsonencode({
    allowed_http_methods = ["POST"]
  })
}

resource "hookdeck_gateway_source_auth" "test" {
  source_id = hookdeck_gateway_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "%[2]s-%[1]s"
  })
}
