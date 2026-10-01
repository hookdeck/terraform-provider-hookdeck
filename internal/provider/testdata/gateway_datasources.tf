
data "hookdeck_gateway_source" "test" {
  id = hookdeck_gateway_source.test.id
}

data "hookdeck_gateway_destination" "test" {
  project_id = %[1]q
  id         = hookdeck_gateway_destination.test.id
}

data "hookdeck_gateway_connection" "test" {
  id = hookdeck_gateway_connection.test.id
}

data "hookdeck_source" "legacy" {
  id = hookdeck_gateway_source.test.id
}
