variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key    = var.hookdeck_org_api_key # hd_org_...
  project_id = "tm_prod"
}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}

# Plan-time error "Project mismatch": this provider manages tm_prod only.
# resource "hookdeck_gateway_source" "other" {
#   project_id = "tm_staging"
#   name       = "orders"
#   type       = "HTTP"
# }
