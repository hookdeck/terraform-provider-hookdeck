terraform {
  required_providers {
    hookdeck = {
      source = "hookdeck/hookdeck"
    }
  }
}

provider "hookdeck" {
  api_base = "https://api.hookdeck.com/2023-07-01"
  api_key  = var.hookdeck_api_key
}

# Create a source
resource "hookdeck_source" "source" {
  # ...
}

# Create a destination
resource "hookdeck_destination" "destination" {
  # ...
}

# Create a connection
resource "hookdeck_connection" "connection" {
  # ...
}
