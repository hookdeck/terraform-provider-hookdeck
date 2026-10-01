package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const (
	providerName = "hookdeck"
	gatewayInfix = "_gateway"
)

// Naming describes one resource type under its v3 name and its v2 alias.
type Naming struct {
	// Suffix is the type name without the provider prefix, e.g. "_source".
	Suffix string
	// Legacy is true for the v2 alias registration.
	Legacy bool
}

func (n Naming) TypeName(providerTypeName string) string {
	if n.Legacy {
		return providerTypeName + n.Suffix
	}
	return providerTypeName + gatewayInfix + n.Suffix
}

func (n Naming) LegacyTypeName() string {
	return providerName + n.Suffix
}

func (n Naming) currentTypeName() string {
	return providerName + gatewayInfix + n.Suffix
}

func (n Naming) ResourceDeprecation() string {
	if !n.Legacy {
		return ""
	}
	return fmt.Sprintf("%s is deprecated and will be removed in v4. Rename it to %s with a moved block. See the v2 to v3 migration guide.", n.LegacyTypeName(), n.currentTypeName())
}

func (n Naming) DataSourceDeprecation() string {
	if !n.Legacy {
		return ""
	}
	return fmt.Sprintf("%s is deprecated and will be removed in v4. Use the %s data source instead. See the v2 to v3 migration guide.", n.LegacyTypeName(), n.currentTypeName())
}

// Description returns description, with a deprecation note on the v2 alias.
func (n Naming) Description(description string) string {
	if !n.Legacy {
		return description
	}
	return fmt.Sprintf("%s. Deprecated: renamed to `%s`; `%s` will be removed in v4.", description, n.currentTypeName(), n.LegacyTypeName())
}

// RenamedStateMover moves state from the v2 alias into the v3 resource.
// Both names share the schema, so the state is copied as-is.
func RenamedStateMover(legacyTypeName string, schema resourceschema.Schema) resource.StateMover {
	return resource.StateMover{
		SourceSchema: &schema,
		StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if req.SourceTypeName != legacyTypeName || !isThisProvider(req.SourceProviderAddress) {
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
				resp.Diagnostics.AddError(
					"Unsupported state for move",
					fmt.Sprintf("%s state could not be decoded. Please report this issue to the provider developers.", legacyTypeName),
				)
				return
			}
			resp.TargetState = *req.SourceState
			resp.TargetPrivate = req.SourcePrivate
		},
	}
}

// isThisProvider matches any registry host and namespace, since mirrors
// report their own. OpenTofu 1.10.0 to 1.12.3 send the bare provider type.
func isThisProvider(address string) bool {
	return address == providerName || strings.HasSuffix(address, "/"+providerName)
}
