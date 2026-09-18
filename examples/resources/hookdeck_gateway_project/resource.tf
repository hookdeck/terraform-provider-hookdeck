# Requires an organization API key with `projects.write`.
resource "hookdeck_gateway_project" "example" {
  name                 = "production"
  notification_methods = ["email"]
}

resource "hookdeck_gateway_source" "example" {
  project_id = hookdeck_gateway_project.example.id
  name       = "example"
  type       = "HTTP"
}
