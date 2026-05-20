resource "hookdeck_destination" "test_%[1]s" {
  name = "test-destination-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url                      = "https://mock.hookdeck.com"
    path_forwarding_disabled = true
  })
}
