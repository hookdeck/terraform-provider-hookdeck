variable "HOOKDECK_API_KEY" {
  type = string
}

variable "HEADER_FILTER_VALUES" {
  type = list(string)
}

terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
    }
  }
}

provider "hookdeck" {
  api_key = var.HOOKDECK_API_KEY
}

resource "hookdeck_source" "my_source" {
  name = "my_source"
}

resource "hookdeck_source_verification" "my_authenticated_source" {
  source_id = hookdeck_source.my_source.id
  verification = {
    basic_auth = {
      username = "example-username"
      password = "example-password"
    }
  }
}

resource "hookdeck_destination" "my_destination" {
  name = "my_destination"
  url  = "https://mock.hookdeck.com"
}

resource "hookdeck_connection" "my_connection" {
  source_id      = hookdeck_source.my_source.id
  destination_id = hookdeck_destination.my_destination.id
  rules = [
    {
      filter_rule = {
        headers = {
          json = jsonencode({
            x-event-type = { "$or" : var.HEADER_FILTER_VALUES }
          })
        }
      }
    }
  ]
}