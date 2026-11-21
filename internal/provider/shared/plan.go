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
// resource, or the config attribute itself) is a real change: the plan is left
// alone, except that the timestamps an update changes are marked unknown. The
// framework marks them before attribute plan modifiers run, so a change that
// only a modifier introduces (config removed from the configuration) would
// leave them known, and the update would then return new values.
func KeepUnchangedPlan(ctx context.Context, state tfsdk.State, plan *tfsdk.Plan) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.Raw.IsNull() || plan.Raw.IsNull() || !plan.Raw.IsKnown() {
		return diags
	}
	if !keepUnchanged(state, plan, &diags) && !diags.HasError() {
		markTimestampsUnknown(plan)
	}
	return diags
}

// updateTimestamps are the computed attributes the API changes on every update.
var updateTimestamps = []string{"updated_at", "disabled_at"}

func markTimestampsUnknown(plan *tfsdk.Plan) {
	var planned map[string]tftypes.Value
	if err := plan.Raw.As(&planned); err != nil {
		return
	}
	attributes := plan.Schema.GetAttributes()
	for _, name := range updateTimestamps {
		attribute, ok := attributes[name]
		if _, inPlan := planned[name]; !ok || !inPlan || !attribute.IsComputed() || attribute.IsOptional() {
			continue
		}
		planned[name] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}
	plan.Raw = tftypes.NewValue(plan.Raw.Type(), planned)
}

// keepUnchanged restores the plan to state when only computed values differ,
// and reports whether it did.
func keepUnchanged(state tfsdk.State, plan *tfsdk.Plan, diags *diag.Diagnostics) bool {
	var planned, prior map[string]tftypes.Value
	if err := plan.Raw.As(&planned); err != nil {
		diags.AddError("Error reading the plan", err.Error())
		return false
	}
	if err := state.Raw.As(&prior); err != nil {
		diags.AddError("Error reading the state", err.Error())
		return false
	}
	restored := make(map[string]tftypes.Value, len(planned))
	for name, value := range planned {
		before, ok := prior[name]
		if !ok {
			return false
		}
		if value.IsFullyKnown() {
			if !value.Equal(before) {
				return false
			}
			restored[name] = value
			continue
		}
		attribute, ok := plan.Schema.GetAttributes()[name]
		if !ok || !attribute.IsComputed() || attribute.IsOptional() {
			return false
		}
		restored[name] = before
	}
	plan.Raw = tftypes.NewValue(plan.Raw.Type(), restored)
	return true
}
