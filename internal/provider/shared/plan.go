package shared

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// KeepUnchangedPlan turns a plan that only differs from state in computed
// values into the state itself. config is Optional and Computed: when the
// configuration sets it, the framework proposes a change to it and marks every
// computed attribute that has no configuration (updated_at, disabled_at) as
// unknown, before the config plan modifier decides there is no change. The
// plan would then show an update for nothing. Call it from ModifyPlan, which
// runs after attribute plan modifiers.
//
// An attribute that is unknown but not only computed (a value from another
// resource, or the config attribute itself) is a real change: the plan is left alone.
func KeepUnchangedPlan(ctx context.Context, state tfsdk.State, plan *tfsdk.Plan) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.Raw.IsNull() || plan.Raw.IsNull() || !plan.Raw.IsKnown() {
		return diags
	}
	var planned, prior map[string]tftypes.Value
	if err := plan.Raw.As(&planned); err != nil {
		diags.AddError("Error reading the plan", err.Error())
		return diags
	}
	if err := state.Raw.As(&prior); err != nil {
		diags.AddError("Error reading the state", err.Error())
		return diags
	}
	restored := make(map[string]tftypes.Value, len(planned))
	for name, value := range planned {
		before, ok := prior[name]
		if !ok {
			return diags
		}
		if value.IsFullyKnown() {
			if !value.Equal(before) {
				return diags
			}
			restored[name] = value
			continue
		}
		attribute, ok := plan.Schema.GetAttributes()[name]
		if !ok || !attribute.IsComputed() || attribute.IsOptional() {
			return diags
		}
		restored[name] = before
	}
	plan.Raw = tftypes.NewValue(plan.Raw.Type(), restored)
	return diags
}
