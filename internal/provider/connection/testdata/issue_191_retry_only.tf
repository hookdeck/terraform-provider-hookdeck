# Exact replica of the configuration from issue #191
# https://github.com/hookdeck/terraform-provider-hookdeck/issues/191
#
# This configuration caused the error:
# "Value Conversion Error"
# "mismatch between struct and object: Struct defines fields not found in object: deduplicate_rule"

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
  depends_on     = [hookdeck_destination.test_%[1]s]
  name           = "test-connection-issue191-%[1]s"
  source_id      = hookdeck_source.test_%[1]s.id
  destination_id = hookdeck_destination.test_%[1]s.id
  rules          = [
    {
      retry_rule = {
        count                 = 6
        interval              = 60000 # 1 minute in milliseconds
        strategy              = "exponential"
        response_status_codes = ["500-599"]
      }
    }
  ]
}
