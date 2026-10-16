resource "hookdeck_source" "test_%[1]s" {
  name = "tf-acc-source-%[1]s"
}

resource "hookdeck_destination" "test_%[1]s" {
  name = "tf-acc-destination-%[1]s"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}

resource "hookdeck_connection" "test_%[1]s" {
  name           = "tf-acc-connection-json-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id

  rules = [
    {
      filter_rule = {
        body = {
          # Using raw JSON string with formatting (like EOT/heredoc would produce)
          json = <<-JSON
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
        }
      }
    }
  ]
}