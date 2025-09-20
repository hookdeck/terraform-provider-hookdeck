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
  name           = "test-connection-json-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id

  rules = [
    {
      filter_rule = {
        body = {
          # Using jsonencode(jsondecode(...)) with heredoc - this should normalize the JSON
          json = jsonencode(jsondecode(<<-JSON
{
  "data": {
    "attributes": {
      "payload": {
        "data": {
          "attributes": {
            "status": {
              "$or": [
                "completed",
                "failed",
                "approved",
                "declined",
                "needs_review"
              ]
            }
          }
        }
      }
    }
  }
}
          JSON
          ))
        }
      }
    }
  ]
}