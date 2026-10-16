data "hookdeck_gateway_project" "by_id" {
  id = %[1]q
}

data "hookdeck_gateway_project" "by_name" {
  name = data.hookdeck_gateway_project.by_id.name
}
