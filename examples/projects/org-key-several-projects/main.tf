variable "hookdeck_org_api_key" {
  type      = string
  sensitive = true
}

provider "hookdeck" {
  api_key = var.hookdeck_org_api_key # hd_org_...
}

resource "hookdeck_gateway_project" "staging" {
  name = "staging"
}

resource "hookdeck_gateway_source" "orders_staging" {
  project_id = hookdeck_gateway_project.staging.id
  name       = "orders"
  type       = "HTTP"
}

# An existing project, looked up by name.
data "hookdeck_gateway_project" "prod" {
  name = "prod"
}

resource "hookdeck_gateway_source" "orders_prod" {
  project_id = data.hookdeck_gateway_project.prod.id
  name       = "orders"
  type       = "HTTP"
}

# Plan-time error "Missing project_id": neither the resource nor the
# provider names a project.
# resource "hookdeck_gateway_source" "missing" {
#   name = "orders"
#   type = "HTTP"
# }
