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
          # Using jsonencode which will produce formatted JSON with newlines/spaces
          json = jsonencode({
            data = {
              attributes = {
                payload = {
                  data = {
                    attributes = {
                      status = {
                        "$or" = [
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
          })
        }
        headers = {
          # Another JSON field to test
          json = jsonencode({
            "x-webhook-type" = "payment.status"
            "x-api-version"  = "v1"
          })
        }
      }
    }
  ]
}