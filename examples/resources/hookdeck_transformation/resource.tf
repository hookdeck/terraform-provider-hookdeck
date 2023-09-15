resource "hookdeck_transformation" "example" {
  name = "example"
  code = file("${path.module}/transformations/transformation_example.js")
  env = jsonencode({
    SECRET = var.secret
  })
}
