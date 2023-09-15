resource "hookdeck_connection" "connection_example" {
  name           = "example"
  description    = "example connection"
  source_id      = hookdeck_source.source_example.id
  destination_id = hookdeck_destination.destination_example.id
  rules = [
    {
      transform_rule = {
        transformation_id = hookdeck_transformation.transformation_example.id
      }
    },
    {
      filter_rule = {
        body = {
          json = file("${path.module}/filter_body.json")
        }
        path = {
          string = "/api/webhook"
        }
      }
    },
    {
      delay_rule = {
        delay = 10000
      }
    },
    {
      retry_rule = {
        count    = 5
        interval = 3600000
        strategy = "exponential"
      }
    }
  ]
}
