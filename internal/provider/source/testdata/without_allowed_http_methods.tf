resource "hookdeck_source" "test_%[1]s" {
  name = "test-source-%[1]s"
  type = "HTTP"
  config = jsonencode({})
}
