package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func outpostSchemaAttributes() map[string]schema.Attribute {
	attributes := schemaAttributes(outpost)
	attributes["config"] = schema.MapAttribute{
		Optional:    true,
		ElementType: types.StringType,
		Description: "Outpost configuration to manage, keyed by Outpost environment variable name (for example `TOPICS` or `RETRY_INTERVAL_SECONDS`). Only the keys set here are managed and tracked; other keys keep their value. Removing a key restores its default. Keys that can carry credentials belong in `sensitive_config`.",
	}
	attributes["sensitive_config"] = schema.MapAttribute{
		Optional:    true,
		Sensitive:   true,
		ElementType: types.StringType,
		Description: "Outpost configuration to manage like `config`, for keys whose values are secrets or can carry credentials, such as `OTEL_EXPORTER_OTLP_HEADERS`, `DESTINATIONS_PROXY_URL` and the `PUBLISH_*` connection strings and secret keys. Values are stored in the Terraform state.",
	}
	return attributes
}

func outpostTimeoutsBlock(ctx context.Context) schema.Block {
	return timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})
}
