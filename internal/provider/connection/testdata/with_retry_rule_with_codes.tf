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
  name           = "tf-acc-connection-retry-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id

  rules = [
    {
      retry_rule = {
        strategy = "exponential"
        count    = 5
        interval = 1000
        response_status_codes = ["500-599"]
      }
    }
  ]
}
