data "hookdeck_outpost_project" "by_id" {
  id = "tm_xxx"
}

data "hookdeck_outpost_project" "by_name" {
  name = "production-outpost"
}

# The whole configuration is readable, including values you did not set.
# config is sensitive; wrap a value that is not a secret in nonsensitive() to output it.
output "max_retry_limit" {
  value = nonsensitive(data.hookdeck_outpost_project.by_name.config["MAX_RETRY_LIMIT"])
}
