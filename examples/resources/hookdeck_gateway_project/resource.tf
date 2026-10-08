# Managing projects requires an organization API key with the
# `projects.write` scope.
resource "hookdeck_gateway_project" "example" {
  name = "production"

  # Deleting a project deletes everything in it.
  lifecycle {
    prevent_destroy = true
  }
}

resource "hookdeck_gateway_source" "example" {
  project_id = hookdeck_gateway_project.example.id
  name       = "example"
  type       = "HTTP"
}
