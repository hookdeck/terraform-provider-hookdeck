resource "hookdeck_gateway_source" "source_example" {
  name = "example"
  type = "HTTP"
}

resource "hookdeck_gateway_destination" "destination_example" {
  name = "example"
  type = "HTTP"
  config = jsonencode({
    url = "https://example.test/webhook"
  })
}

resource "hookdeck_gateway_transformation" "transformation_example" {
  name = "example"
  code = "addHandler('transform', (request, context) => request);"
}

resource "hookdeck_gateway_connection" "connection_example" {
  name           = "example"
  description    = "example connection"
  source_id      = hookdeck_gateway_source.source_example.id
  destination_id = hookdeck_gateway_destination.destination_example.id
  rules = [
    {
      transform_rule = {
        transformation_id = hookdeck_gateway_transformation.transformation_example.id
      }
    },
    {
      filter_rule = {
        body = {
          # you can use a file for the filter JSON
          json = jsonencode(jsondecode(file("${path.module}/filter_body.json")))
        }
        headers = {
          # or use Terraform's `jsonencode` to inline the JSON
          json = jsonencode({
            authorization = "my_super_secret_key"
          })
        }
        path = {
          # or match with a `string`, `number`, or `boolean` value
          string = "/api/webhook"
        }
      }
    },
    {
      deduplicate_rule = {
        window         = 3600000 # 1 hour in milliseconds (1s-1hr range)
        include_fields = ["body.id", "body.event_type"]
      }
    },
    {
      delay_rule = {
        delay = 10000
      }
    },
    {
      retry_rule = {
        count                 = 5
        interval              = 3600000
        strategy              = "exponential"
        response_status_codes = ["500-599", ">400", "!503"]
      }
    }
  ]
}
