provider "hookdeck" {
  project_id = %[2]q
}

resource "hookdeck_gateway_source" "test" {
  name = "v3-src-%[1]s"
  type = "HTTP"
}
