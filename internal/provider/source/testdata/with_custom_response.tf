resource "hookdeck_source" "test_%[1]s" {
  name = "test-source-%[1]s"
  type = "HTTP"
  config = jsonencode({
    custom_response = {
      content_type = "json"
      body         = "{\"ok\":true}"
    }
  })
}
