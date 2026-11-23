package project

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type outpostProjectModel struct {
	projectResourceModel
	Config          types.Map      `tfsdk:"config"`
	SensitiveConfig types.Map      `tfsdk:"sensitive_config"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

type outpostProjectDataModel struct {
	projectResourceModel
	Config types.Map `tfsdk:"config"`
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

// setManaged writes the live values back to the map that holds each key. A
// map keeps only the keys it already has, so state holds what the user set.
func (m *outpostProjectModel) setManaged(ctx context.Context, merged map[string]string) diag.Diagnostics {
	var diags diag.Diagnostics
	narrow := func(prior types.Map) types.Map {
		own, d := stringMap(ctx, prior)
		diags.Append(d...)
		kept := map[string]string{}
		for k := range own {
			if v, ok := merged[k]; ok {
				kept[k] = v
			}
		}
		if len(kept) == 0 && prior.IsNull() {
			return prior
		}
		value, d := types.MapValueFrom(ctx, types.StringType, kept)
		diags.Append(d...)
		return value
	}
	m.Config = narrow(m.Config)
	m.SensitiveConfig = narrow(m.SensitiveConfig)
	return diags
}

func toPatchValues(values map[string]string) map[string]*string {
	out := make(map[string]*string, len(values))
	for k, v := range values {
		out[k] = &v
	}
	return out
}
