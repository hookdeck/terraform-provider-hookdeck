resource "hookdeck_source" "example" {
  name                 = "example"
  description          = "example source"
  allowed_http_methods = ["GET", "POST", "PUT"]
  custom_response = {
    content_type = "json"
    body         = "{\"hello\": \"world\"}"
  }
}
