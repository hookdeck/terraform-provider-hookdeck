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
		Description: "Outpost configuration to manage, keyed by Outpost environment variable name (for example `TOPICS` or `RETRY_INTERVAL_SECONDS`). Only the keys set here and in `sensitive_config` are managed and tracked; other keys keep their value. Removing a key restores its default. A key goes in one of the two maps.",
	}
	attributes["sensitive_config"] = schema.MapAttribute{
		Optional:    true,
		Sensitive:   true,
		ElementType: types.StringType,
		Description: "Outpost configuration to manage like `config`, hidden from plan output. Put the values you do not want shown here; any key can go in either map. Values are stored in the Terraform state.",
	}
	return attributes
}

func outpostTimeoutsBlock(ctx context.Context) schema.Block {
	return timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})
}
