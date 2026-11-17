resource "hookdeck_gateway_source" "test" {
  name = "tf-acc-v3-src-%[1]s"
  type = "HTTP"
  config = jsonencode({
    allowed_http_methods = ["POST"]
  })
}

resource "hookdeck_gateway_destination" "test" {
  name = "tf-acc-v3-dst-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}
