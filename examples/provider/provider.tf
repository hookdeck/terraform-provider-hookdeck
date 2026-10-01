terraform {
  required_providers {
    hookdeck = {
      source  = "hookdeck/hookdeck"
      version = "~> 3.0"
    }
  }
}

variable "hookdeck_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_api_key
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}

resource "hookdeck_gateway_destination" "api" {
  name = "api"
  type = "HTTP"
  config = jsonencode({
    url = "https://api.example.com/webhooks"
  })
}

resource "hookdeck_gateway_connection" "orders_to_api" {
  source_id      = hookdeck_gateway_source.orders.id
  destination_id = hookdeck_gateway_destination.api.id
}
