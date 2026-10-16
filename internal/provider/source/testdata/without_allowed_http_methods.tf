resource "hookdeck_source" "test_%[1]s" {
  name = "tf-acc-source-%[1]s"
  type = "HTTP"
  config = jsonencode({})
}
