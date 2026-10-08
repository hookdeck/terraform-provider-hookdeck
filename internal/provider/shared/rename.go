package shared

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

// AsLegacy returns the naming of the v2 alias of the same type.
func (n Naming) AsLegacy() Naming {
	n.Legacy = true
	return n
}

// WithTeamID adds team_id to the attributes of a v2 name. The v3 names
// have project_id only.
func (n Naming) WithTeamID(attributes map[string]resourceschema.Attribute) map[string]resourceschema.Attribute {
	if n.Legacy {
		attributes[teamIDAttribute] = resourceschema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Description: "ID of the project the resource belongs to. Same value as `project_id`.",
		}
	}
	return attributes
}

// RenamedStateMover moves state from the v2 alias into the v3 resource.
// Every attribute is copied; team_id is dropped, and fills project_id in
// state written before project_id existed.
func RenamedStateMover(legacyTypeName string, source, target resourceschema.Schema) resource.StateMover {
	return resource.StateMover{
		SourceSchema: &source,
		StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
			if req.SourceTypeName != legacyTypeName || !isThisProvider(req.SourceProviderAddress) {
				return
			}
			if req.SourceSchemaVersion != target.Version {
				resp.Diagnostics.AddError(
					"Unsupported state version for move",
					fmt.Sprintf("%s state is at schema version %d, this provider expects %d. Run terraform apply under the old name first so the state is upgraded, then move it.", legacyTypeName, req.SourceSchemaVersion, target.Version),
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
			raw, err := renamedState(ctx, req.SourceState.Raw, target)
			if err != nil {
				resp.Diagnostics.AddError(
					"Unsupported state for move",
					fmt.Sprintf("%s state could not be moved: %s. Please report this issue to the provider developers.", legacyTypeName, err),
				)
				return
			}
			resp.TargetState.Raw = raw
			resp.TargetPrivate = req.SourcePrivate
		},
	}
}

func renamedState(ctx context.Context, source tftypes.Value, target resourceschema.Schema) (tftypes.Value, error) {
	var attributes map[string]tftypes.Value
	if err := source.As(&attributes); err != nil {
		return tftypes.Value{}, err
	}
	if teamID, ok := attributes[teamIDAttribute]; ok {
		if projectID, ok := attributes[projectIDAttribute]; !ok || projectID.IsNull() {
			attributes[projectIDAttribute] = teamID
		}
	}
	targetType, ok := target.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("target schema type is %s, not an object", target.Type().TerraformType(ctx))
	}
	moved := make(map[string]tftypes.Value, len(targetType.AttributeTypes))
	for name, typ := range targetType.AttributeTypes {
		value, ok := attributes[name]
		if !ok {
			value = tftypes.NewValue(typ, nil)
		}
		moved[name] = value
	}
	return tftypes.NewValue(targetType, moved), nil
}

// isThisProvider matches any registry host and namespace, since mirrors
// report their own. OpenTofu 1.10.0 to 1.12.3 send the bare provider type.
func isThisProvider(address string) bool {
	return address == providerName || strings.HasSuffix(address, "/"+providerName)
}

// ModelSource is a plan, state or configuration a model is read from.
type ModelSource interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

// ModelTarget is a state a model is written to.
type ModelTarget interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}
