resource "hookdeck_destination" {
  name = "test-destination-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
    delivery_policy = {
      period = "concurrent"
    }
  })
}
