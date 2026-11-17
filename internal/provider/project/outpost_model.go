package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outpostProjectModel struct {
	projectResourceModel
	Config                   types.Map      `tfsdk:"config"`
	SensitiveConfig          types.Map      `tfsdk:"sensitive_config"`
	PortalCustomDomain       types.String   `tfsdk:"portal_custom_domain"`
	PortalCustomDomainStatus types.Object   `tfsdk:"portal_custom_domain_status"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

type outpostProjectDataModel struct {
	projectResourceModel
	Config                   types.Map    `tfsdk:"config"`
	SensitiveConfig          types.Map    `tfsdk:"sensitive_config"`
	PortalCustomDomain       types.String `tfsdk:"portal_custom_domain"`
	PortalCustomDomainStatus types.Object `tfsdk:"portal_custom_domain_status"`
}

func stringMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	out := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return out, nil
	}
	return out, m.ElementsAs(ctx, &out, false)
}

func (m *outpostProjectModel) managed(ctx context.Context) (map[string]string, diag.Diagnostics) {
	merged, diags := stringMap(ctx, m.Config)
	sensitive, d := stringMap(ctx, m.SensitiveConfig)
	diags.Append(d...)
	for k, v := range sensitive {
		merged[k] = v
	}
	return merged, diags
}

const redactedValue = "(redacted: URL with credentials)"

// setManaged writes the merged values back to the map that owns each key.
// A value that carries credentials is never written to config: state holds
// redactedValue for it, so the plan shows the change without the value.
func (m *outpostProjectModel) setManaged(ctx context.Context, merged map[string]string) diag.Diagnostics {
	var diags diag.Diagnostics
	narrow := func(prior types.Map, at path.Path, plain bool) types.Map {
		own, d := stringMap(ctx, prior)
		diags.Append(d...)
		kept := map[string]string{}
		for _, k := range sortedKeys(own) {
			v, ok := merged[k]
			if !ok {
				continue
			}
			if plain && carriesCredentials(k, v) {
				diags.AddAttributeWarning(at.AtMapKey(k), "Credentials in config",
					fmt.Sprintf("The value of %s in Hookdeck is a URL that carries credentials, so it is not stored in config. Applying sets %s to the configured value. To keep the value Hookdeck has, set the key in sensitive_config instead.", k, k))
				v = redactedValue
			}
			kept[k] = v
		}
		if len(kept) == 0 && prior.IsNull() {
			return prior
		}
		value, d := types.MapValueFrom(ctx, types.StringType, kept)
		diags.Append(d...)
		return value
	}
	m.Config = narrow(m.Config, path.Root("config"), true)
	m.SensitiveConfig = narrow(m.SensitiveConfig, path.Root("sensitive_config"), false)
	return diags
}

func toPatchValues(values map[string]string) map[string]*string {
	out := make(map[string]*string, len(values))
	for k, v := range values {
		out[k] = &v
	}
	return out
}
