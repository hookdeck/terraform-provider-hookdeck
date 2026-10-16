resource "hookdeck_gateway_source" "example" {
  name = "example"
  type = "HTTP"
}

resource "hookdeck_gateway_source_auth" "example" {
  source_id = hookdeck_gateway_source.example.id
  auth_type = "BASIC_AUTH"
  auth = jsonencode({
    username = "username"
    password = "password"
  })
}
