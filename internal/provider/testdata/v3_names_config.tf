resource "hookdeck_gateway_source" "test" {
  name = "tf-acc-v3-src-%[1]s"
  type = "HTTP"
  config = jsonencode({
    allowed_http_methods = ["POST"]
    custom_response = {
      content_type = "json"
      body         = "{}"
    }
  })
}

resource "hookdeck_gateway_destination" "test" {
  name = "tf-acc-v3-dst-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url       = "https://mock.hookdeck.com"
    auth_type = "BEARER_TOKEN"
    auth = {
      token = "tf-acc-token"
    }
    delivery_policy = {
      rate   = "10"
      period = "second"
    }
  })
}
