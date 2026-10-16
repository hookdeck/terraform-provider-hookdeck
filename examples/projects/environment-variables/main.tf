# export HOOKDECK_API_KEY=...          a project or an organization API key
# export HOOKDECK_PROJECT_ID=tm_prod   optional with a project API key
provider "hookdeck" {}

resource "hookdeck_gateway_source" "orders" {
  name = "orders"
  type = "HTTP"
}
