resource "hookdeck_source" "test" {
  name = "v2-src-%[1]s"
  type = "HTTP"
}

resource "hookdeck_destination" "test" {
  name = "v2-dst-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}

resource "hookdeck_transformation" "test" {
  name = "v2-trs-%[1]s"
  code = "exports.handler = async (request, context) => { return request; };"
}

resource "hookdeck_connection" "test" {
  name           = "v2-con-%[1]s"
  source_id      = hookdeck_source.test.id
  destination_id = hookdeck_destination.test.id
  rules = [{
    transform_rule = {
      transformation_id = hookdeck_transformation.test.id
    }
  }]
}
