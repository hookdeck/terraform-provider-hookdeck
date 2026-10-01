
resource "hookdeck_gateway_source" "test" {
  project_id = hookdeck_gateway_project.test.id
  name       = "v3-src-%[1]s"
  type       = "HTTP"
}
