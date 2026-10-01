
resource "hookdeck_gateway_source" "other" {
  project_id = %[2]q
  name       = "v3-src-other-%[1]s"
  type       = "HTTP"
}
