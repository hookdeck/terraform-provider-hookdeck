variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  alias      = "prod"
  api_key    = var.hookdeck_org_api_key
  project_id = "tm_prod"
}

provider "hookdeck" {
  alias      = "staging"
  api_key    = var.hookdeck_org_api_key
  project_id = "tm_staging"
}

resource "hookdeck_gateway_source" "orders_prod" {
  provider = hookdeck.prod
  name     = "orders"
  type     = "HTTP"
}

resource "hookdeck_gateway_source" "orders_staging" {
  provider = hookdeck.staging
  name     = "orders"
  type     = "HTTP"
}
