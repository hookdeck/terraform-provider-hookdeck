resource "hookdeck_source" "test" {
%[1]s  name = "v2-src"
}

resource "hookdeck_source_auth" "test" {
%[1]s  source_id = hookdeck_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "secret"
  })
}
