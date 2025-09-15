resource "hookdeck_transformation" "example" {
  name = "example"
  code = file("${path.module}/transformations/transformation_example.js")
  env = jsonencode({
    SECRET = var.secret
  })

  # Important: Use create_before_destroy to ensure proper deletion order
  # when the transformation is referenced by connections
  lifecycle {
    create_before_destroy = true
  }
}
