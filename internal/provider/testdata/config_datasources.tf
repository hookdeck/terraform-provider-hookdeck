resource "hookdeck_gateway_source" "test" {
  name = "tf-acc-cfg-src-%[1]s"
  type = "HTTP"
  config = jsonencode({
    allowed_http_methods = ["POST"]
  })
}

resource "hookdeck_gateway_destination" "test" {
  name = "tf-acc-cfg-dst-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}

data "hookdeck_gateway_source" "test" {
  id = hookdeck_gateway_source.test.id
}

data "hookdeck_gateway_destination" "test" {
  id = hookdeck_gateway_destination.test.id
}

output "source_config" {
  value     = jsondecode(data.hookdeck_gateway_source.test.config)
  sensitive = true
}

output "destination_config" {
  value     = jsondecode(data.hookdeck_gateway_destination.test.config)
  sensitive = true
}
