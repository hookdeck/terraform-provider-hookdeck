variable "otlp_headers" {
  type      = string
  sensitive = true
}

resource "hookdeck_outpost_project" "example" {
  name = "production-outpost"

  # Outpost environment variables to manage. Keys you leave out keep their value.
  config = {
    TOPICS                      = "order.created,order.updated"
    RETRY_INTERVAL_SECONDS      = "60"
    PORTAL_ORGANIZATION_NAME    = "Acme"
    OTEL_SERVICE_NAME           = "acme-outpost"
    OTEL_EXPORTER_OTLP_ENDPOINT = "https://otel.example.com"
  }

  # Same as config, but the values are hidden in plan output.
  sensitive_config = {
    OTEL_EXPORTER_OTLP_HEADERS = var.otlp_headers
  }

  # Deleting a project deletes everything in it.
  # lifecycle {
  #   prevent_destroy = true
  # }
}
