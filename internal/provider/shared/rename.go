package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GatewayPrefix is the type-name prefix of Event Gateway resources.
const GatewayPrefix = "_gateway"

// Naming describes one resource type under its v3 name and its v2 alias.
type Naming struct {
	// Suffix is the type name without the provider prefix, e.g. "_source".
	Suffix string
	// Legacy is true for the v2 alias registration.
	Legacy bool
}

// TypeName returns the full type name for the provider type name.
func (n Naming) TypeName(providerTypeName string) string {
	if n.Legacy {
		return providerTypeName + n.Suffix
	}
	return providerTypeName + GatewayPrefix + n.Suffix
}

// LegacyTypeName is the v2 name this resource moves from.
func (n Naming) LegacyTypeName() string {
	return "hookdeck" + n.Suffix
}

// DeprecationMessage is set on the v2 alias schema.
func (n Naming) DeprecationMessage() string {
	if !n.Legacy {
		return ""
	}
	return fmt.Sprintf("hookdeck%s is deprecated and will be removed in v4. Rename to hookdeck%s%s with a moved block; see the v3 upgrade guide.", n.Suffix, GatewayPrefix, n.Suffix)
}

// RenamedStateMover moves state from the v2 alias into the v3 resource.
// Both names share the schema, so the state is copied as-is.
func RenamedStateMover(legacyTypeName string, schema resourceschema.Schema) resource.StateMover {
	return resource.StateMover{
		SourceSchema: &schema,
		StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if req.SourceTypeName != legacyTypeName || !strings.HasSuffix(req.SourceProviderAddress, "/hookdeck") {
				return
			}
			if req.SourceSchemaVersion != schema.Version {
				resp.Diagnostics.AddError(
					"Unsupported state version for move",
					fmt.Sprintf("%s state is at schema version %d, this provider expects %d. Run terraform apply under the old name first so the state is upgraded, then move it.", legacyTypeName, req.SourceSchemaVersion, schema.Version),
				)
				return
			}
			if req.SourceState == nil {
				resp.Diagnostics.AddError("Unsupported state for move", fmt.Sprintf("%s state could not be decoded.", legacyTypeName))
				return
			}
			resp.TargetState = *req.SourceState
			resp.TargetPrivate = req.SourcePrivate
		},
	}
}
