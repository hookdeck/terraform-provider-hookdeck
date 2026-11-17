data "hookdeck_outpost_project" "by_id" {
  id = "tm_xxx"
}

data "hookdeck_outpost_project" "by_name" {
  name = "production-outpost"
}

# The whole configuration is readable, including values you did not set.
output "max_retry_limit" {
  value = data.hookdeck_outpost_project.by_name.config["MAX_RETRY_LIMIT"]
}
