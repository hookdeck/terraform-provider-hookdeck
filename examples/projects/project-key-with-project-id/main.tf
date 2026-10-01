variable "hookdeck_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key    = var.hookdeck_api_key # project API key
  project_id = "tm_prod"
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}
