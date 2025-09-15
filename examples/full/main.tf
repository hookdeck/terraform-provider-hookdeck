variable "HOOKDECK_API_KEY" {
  type = string
}

variable "HEADER_FILTER_VALUES" {
  type = list(string)
}

variable "STRIPE_SECRET_KEY" {
  type = string
}

variable "EXISTING_SOURCE_ID" {
  type    = string
  default = null
}

variable "EXISTING_DESTINATION_ID" {
  type    = string
  default = null
}

variable "EXISTING_CONNECTION_ID" {
  type    = string
  default = null
}

terraform {
  required_providers {
    hookdeck = {
      source = "hookdeck/hookdeck"
    }
  }
}

provider "hookdeck" {
  api_key = var.HOOKDECK_API_KEY
}

resource "hookdeck_source" "standalone_source" {
  name = "untyped_source"
}

resource "hookdeck_source" "first_source" {
  name = "first_source"
  type = "HTTP"
  config = jsonencode({
    auth_type = "BASIC_AUTH"
    auth = {
      username = "example-username"
      password = "example-password"
    }
  })
}

resource "hookdeck_source" "second_source" {
  name = "second_source"
  type = "HTTP"
  config = jsonencode({
    url       = "https://mock.hookdeck.com/test"
    auth_type = "BASIC_AUTH"
    auth = {
      username = "some-username"
      password = "blah-blah-blah"
    }
  })
}

resource "hookdeck_source" "third_source" {
  name = "third_source"
  type = "HTTP"
}

resource "hookdeck_source_auth" "third_source_auth" {
  source_id = hookdeck_source.third_source.id
  auth_type = "BASIC_AUTH"
  auth = jsonencode({
    username = "some-username"
    password = "blah-blah-blah"
  })
}

resource "hookdeck_destination" "first_destination" {
  name = "first_destination"
  type = "MOCK_API"
}

resource "hookdeck_destination" "second_destination" {
  name = "second_destination"
  type = "HTTP"
  config = jsonencode({
    url       = "https://mock.hookdeck.com"
    auth_type = "BASIC_AUTH"
    auth = {
      username = "username"
      password = "password"
    }
    rate_limit        = 10
    rate_limit_period = "concurrent"
  })
}

resource "hookdeck_destination" "aws_destination" {
  name = "aws_destination"
  config = jsonencode({
    url       = "https://mock.hookdeck.com"
    auth_type = "AWS_SIGNATURE"
    auth = {
      access_key_id     = "some-access"
      secret_access_key = "some-secret"
      region            = "us-west-2"
      service           = "lambda"
    }
  })
}

resource "hookdeck_transformation" "example_transformation" {
  name = "example_transformation"
  code = <<EOT
addHandler("transform", (request, context) => {
  request.headers["example-header"] = "Hello World";
  return request;
});
EOT

  # Important: Use create_before_destroy to ensure proper deletion order
  # when the transformation is referenced by connections
  lifecycle {
    create_before_destroy = true
  }
}

resource "hookdeck_connection" "first_connection" {
  source_id      = hookdeck_source.first_source.id
  destination_id = hookdeck_destination.first_destination.id
  rules = [
    {
      transform_rule = {
        transformation_id = hookdeck_transformation.example_transformation.id
      }
    },
    {
      filter_rule = {
        headers = {
          json = jsonencode({
            x-event-type = { "$or" : var.HEADER_FILTER_VALUES }
          })
        }
      }
    },
  ]
}

resource "hookdeck_connection" "second_connection" {
  source_id      = hookdeck_source.second_source.id
  destination_id = hookdeck_destination.first_destination.id
}

data "hookdeck_source" "manually_created_source" {
  id = var.EXISTING_SOURCE_ID
}

data "hookdeck_destination" "manually_created_destination" {
  id = var.EXISTING_DESTINATION_ID
}

data "hookdeck_connection" "manually_created_connection" {
  id = var.EXISTING_CONNECTION_ID
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

resource "hookdeck_source" "stripe_source" {
  name = "stripe"
  type = "STRIPE"
}

resource "hookdeck_webhook_registration" "stripe_registration" {
  provider = hookdeck

  register = {
    request = {
      method = "POST"
      url    = "https://api.stripe.com/v1/webhook_endpoints"
      headers = jsonencode({
        authorization = "Bearer ${var.STRIPE_SECRET_KEY}"
      })
      body = "url=${hookdeck_source.stripe_source.url}&enabled_events[]=charge.failed&enabled_events[]=charge.succeeded"
    }
  }
  unregister = {
    request = {
      method = "DELETE"
      url    = "https://api.stripe.com/v1/webhook_endpoints/{{.register.response.body.id}}"
      headers = jsonencode({
        authorization = "Bearer ${var.STRIPE_SECRET_KEY}"
      })
    }
  }
}

resource "hookdeck_source_auth" "stripe_source_auth" {
  source_id = hookdeck_source.stripe_source.id
  auth = jsonencode({
    webhook_secret_key = jsondecode(hookdeck_webhook_registration.stripe_registration.register.response).body.secret
  })
}
