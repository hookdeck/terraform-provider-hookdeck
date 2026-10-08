resource "hookdeck_gateway_source" "test" {
  name = "v2-src-%[1]s"
  type = "HTTP"
}

resource "hookdeck_gateway_source_auth" "test" {
  source_id = hookdeck_gateway_source.test.id
  auth_type = "API_KEY"
  auth = jsonencode({
    header_key = "x-api-key"
    api_key    = "secret-%[1]s"
  })
}

resource "hookdeck_gateway_destination" "test" {
  name = "v2-dst-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}

resource "hookdeck_gateway_transformation" "test" {
  name = "v2-trs-%[1]s"
  code = "exports.handler = async (request, context) => { return request; };"
}

resource "hookdeck_gateway_connection" "test" {
  name           = "v2-con-%[1]s"
  source_id      = hookdeck_gateway_source.test.id
  destination_id = hookdeck_gateway_destination.test.id
  rules = [{
    transform_rule = {
      transformation_id = hookdeck_gateway_transformation.test.id
    }
  }]
}

moved {
  from = hookdeck_source.test
  to   = hookdeck_gateway_source.test
}

moved {
  from = hookdeck_source_auth.test
  to   = hookdeck_gateway_source_auth.test
}

moved {
  from = hookdeck_destination.test
  to   = hookdeck_gateway_destination.test
}

moved {
  from = hookdeck_transformation.test
  to   = hookdeck_gateway_transformation.test
}

moved {
  from = hookdeck_connection.test
  to   = hookdeck_gateway_connection.test
}
