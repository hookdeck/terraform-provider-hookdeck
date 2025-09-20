resource "hookdeck_transformation" "test_%[1]s" {
  name = "test-transformation-env-%[1]s"
  code = "exports.handler = async (request, context) => { return { ...request, apiKey: context.env.API_KEY }; };"
  env = jsonencode({
    API_KEY = "test-key-%[1]s"
    DEBUG   = "true"
  })
}
