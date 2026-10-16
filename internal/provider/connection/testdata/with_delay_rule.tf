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
  name           = "tf-acc-connection-delay-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id

  rules = [
    {
      delay_rule = {
        delay = 5000
      }
    }
  ]
}
