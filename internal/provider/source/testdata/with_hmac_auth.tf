resource "hookdeck_source" "test_%[1]s" {
  name = "test-source-%[1]s"
  type = "HTTP"
  config = jsonencode({
    auth_type = "HMAC"
    auth = {
      algorithm          = "sha256"
      encoding           = "base64"
      header_key         = "X-Probe-Sig"
      webhook_secret_key = "probe-secret-%[1]s"
    }
  })
}
