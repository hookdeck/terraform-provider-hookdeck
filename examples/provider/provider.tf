terraform {
  required_providers {
    hookdeck = {
      source = "hookdeck/hookdeck"
    }
  }
}

# Project API key: every resource belongs to the key's project.
provider "hookdeck" {
  api_key = var.hookdeck_api_key
}

# Organization API key: set a default project, or project_id per resource.
# provider "hookdeck" {
#   api_key    = var.hookdeck_org_api_key
#   project_id = "tm_xxx"
# }

# Create a source
resource "hookdeck_gateway_source" "source" {
  # ...
}

# Create a destination
resource "hookdeck_gateway_destination" "destination" {
  # ...
}

# Create a connection
resource "hookdeck_gateway_connection" "connection" {
  # ...
}
