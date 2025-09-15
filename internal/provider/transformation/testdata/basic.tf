resource "hookdeck_transformation" "test_%[1]s" {
  name = "test-transformation-%[1]s"
  code = "exports.handler = async (request, context) => { return request; };"
}
