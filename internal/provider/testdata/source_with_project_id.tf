resource "hookdeck_gateway_source" "test" {
  project_id = %[2]q
  name       = "v3-src-%[1]s"
  type       = "HTTP"
}
