# A source auth is imported by the ID of its source.
$ terraform import hookdeck_gateway_source_auth.example <source_id>

# With an organization API key and no provider project_id, name the project:
$ terraform import hookdeck_gateway_source_auth.example <project_id>/<source_id>
