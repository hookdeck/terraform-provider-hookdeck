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
      version = "0.5.0-beta.1"
    }
  }
}

provider "hookdeck" {
  api_key = var.HOOKDECK_API_KEY
}

resource "hookdeck_source" "first_source" {
  name = "first_source"
}

resource "hookdeck_source" "second_source" {
  name = "my_source"
}

resource "hookdeck_source_verification" "named_basic_auth_verification" {
  source_id = hookdeck_source.first_source.id
  verification = {
    basic_auth = {
      username = "example-username"
      password = "example-password"
    }
  }
}

resource "hookdeck_source_verification" "json_basic_auth_verification" {
  source_id = hookdeck_source.second_source.id
  verification = {
    json = jsonencode({
      type = "basic_auth"
      configs = {
        username = "some-username"
        password = "blah-blah-blah"
      }
    })
  }
}

resource "hookdeck_destination" "first_destination" {
  name = "first_destination"
  url  = "https://mock.hookdeck.com"
}

resource "hookdeck_destination" "second_destination" {
  name = "second_destination"
  url  = "https://mock.hookdeck.com"
  auth_method = {
    basic_auth = {
      username = "some-username"
      password = "blah-blah-blah"
    }
  }
  rate_limit = {
    period = "concurrent"
    limit = 10
  }
}

resource "hookdeck_destination" "aws_destination" {
  name = "aws_destination"
  url  = "https://mock.hookdeck.com"
  auth_method = {
    aws_signature = {
      access_key_id     = "some-access"
      secret_access_key = "some-secret"
      region            = "us-west-2"
      service           = "lambda"
    }
  }
}

resource "hookdeck_connection" "first_connection" {
  source_id      = hookdeck_source.first_source.id
  destination_id = hookdeck_destination.first_destination.id
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

resource "hookdeck_connection" "second_connection" {
  source_id      = hookdeck_source.second_source.id
  destination_id = hookdeck_destination.first_destination.id
}

data "hookdeck_source" "manually_created_source" {
  id = "src_112rkwa855tb0z"
}

data "hookdeck_destination" "manually_created_destination" {
  id = "des_tsrZIbyk0JBB"
}

data "hookdeck_connection" "manually_created_connection" {
  id = "web_xDRnu9yq9GMl"
}

resource "hookdeck_connection" "first_connection_using_data_sources" {
  name           = "first_connection_using_data_sources"
  source_id      = data.hookdeck_source.manually_created_source.id
  destination_id = data.hookdeck_destination.manually_created_destination.id
}

resource "hookdeck_connection" "second_connection_using_data_sources" {
  name           = "second_connection_using_data_sources"
  source_id      = data.hookdeck_connection.manually_created_connection.source_id
  destination_id = data.hookdeck_connection.manually_created_connection.destination_id
}