resource "hookdeck_source_auth" "example" {
  source_id = hookdeck_source.example.id
  auth_type = "BASIC_AUTH"
  auth = jsonencode({
    username = "username"
    password = "password"
  })
}
