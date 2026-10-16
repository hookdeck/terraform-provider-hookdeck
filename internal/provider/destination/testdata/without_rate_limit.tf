resource "hookdeck_destination" "test_%[1]s" {
  name = "tf-acc-destination-%[1]s"
  type = "HTTP"
  config = jsonencode({
    url = "https://mock.hookdeck.com"
  })
}
