resource "hookdeck_transformation" "test_%[1]s" {
  name = "test-transformation-%[1]s"
  code = "exports.handler = async (request, context) => { return request; };"
}

resource "hookdeck_source" "test_%[1]s" {
  name = "test-source-%[1]s"
}

resource "hookdeck_destination" "test_%[1]s" {
  name = "test-destination-%[1]s"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}

resource "hookdeck_connection" "test_%[1]s" {
  name           = "test-connection-reordered-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id

  rules = [
    {
      retry_rule = {
        strategy = "exponential"
        count    = 3
        interval = 5000
      }
    },
    {
      transform_rule = {
        transformation_id = hookdeck_transformation.test_%[1]s.id
      }
    },
    {
      delay_rule = {
        delay = 2000
      }
    },
    {
      filter_rule = {
        headers = {
          json = jsonencode({
            "x-webhook-type" = "order.created"
          })
        }
      }
    }
  ]
}
