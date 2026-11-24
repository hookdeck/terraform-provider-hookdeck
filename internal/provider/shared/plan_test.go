package shared

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestKeepUnchangedPlan(t *testing.T) {
	ctx := context.Background()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true},
		"name":       schema.StringAttribute{Required: true},
		"type":       schema.StringAttribute{Required: true},
		"project_id": schema.StringAttribute{Optional: true, Computed: true},
		"config":     schema.StringAttribute{Optional: true, Computed: true},
		"updated_at": schema.StringAttribute{Computed: true},
	}}
	objectType := s.Type().TerraformType(ctx)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	str := func(v string) tftypes.Value { return tftypes.NewValue(tftypes.String, v) }
	object := func(overrides map[string]tftypes.Value) tftypes.Value {
		values := map[string]tftypes.Value{
			"id":         str("src_1"),
			"name":       str("a"),
			"type":       str("HTTP"),
			"project_id": str("tm_1"),
			"config":     str(`{"url":"x","extra":1}`),
			"updated_at": str("t0"),
		}
		for k, v := range overrides {
			values[k] = v
		}
		return tftypes.NewValue(objectType, values)
	}
	state := object(nil)

	tests := []struct {
		name string
		plan tftypes.Value
		// want is the plan after the call: state when restored, plan when left alone.
		restored bool
	}{
		{name: "computed values unknown and nothing else differs", plan: object(map[string]tftypes.Value{"updated_at": unknown}), restored: true},
		{name: "several computed values unknown", plan: object(map[string]tftypes.Value{"updated_at": unknown, "id": unknown}), restored: true},
		{name: "config planned equal to state", plan: object(map[string]tftypes.Value{"config": str(`{"url":"x","extra":1}`), "updated_at": unknown}), restored: true},
		{name: "a known attribute changed", plan: object(map[string]tftypes.Value{"name": str("b"), "updated_at": unknown})},
		{name: "config planned differently", plan: object(map[string]tftypes.Value{"config": str(`{"url":"y"}`), "updated_at": unknown})},
		{name: "type changed", plan: object(map[string]tftypes.Value{"type": str("STRIPE"), "updated_at": unknown})},
		{name: "project changed", plan: object(map[string]tftypes.Value{"project_id": str("tm_2"), "updated_at": unknown})},
		{name: "config unknown", plan: object(map[string]tftypes.Value{"config": unknown, "updated_at": unknown})},
		{name: "value from another resource unknown", plan: object(map[string]tftypes.Value{"name": unknown, "updated_at": unknown})},
		{name: "optional value from another resource unknown", plan: object(map[string]tftypes.Value{"project_id": unknown, "updated_at": unknown})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := tfsdk.Plan{Schema: s, Raw: tt.plan}
			if d := KeepUnchangedPlan(ctx, tfsdk.State{Schema: s, Raw: state}, &plan); d.HasError() {
				t.Fatal(d)
			}
			want := tt.plan
			if tt.restored {
				want = state
			}
			if !plan.Raw.Equal(want) {
				t.Fatalf("plan = %s\nwant %s", plan.Raw, want)
			}
		})
	}

	t.Run("config unknown marks the timestamps unknown", func(t *testing.T) {
		plan := tfsdk.Plan{Schema: s, Raw: object(map[string]tftypes.Value{"config": unknown})}
		if d := KeepUnchangedPlan(ctx, tfsdk.State{Schema: s, Raw: state}, &plan); d.HasError() {
			t.Fatal(d)
		}
		want := object(map[string]tftypes.Value{"config": unknown, "updated_at": unknown})
		if !plan.Raw.Equal(want) {
			t.Fatalf("plan = %s\nwant %s", plan.Raw, want)
		}
	})
	t.Run("create", func(t *testing.T) {
		plan := tfsdk.Plan{Schema: s, Raw: object(map[string]tftypes.Value{"id": unknown, "updated_at": unknown})}
		before := plan.Raw.Copy()
		if d := KeepUnchangedPlan(ctx, tfsdk.State{Schema: s, Raw: tftypes.NewValue(objectType, nil)}, &plan); d.HasError() {
			t.Fatal(d)
		}
		if !plan.Raw.Equal(before) {
			t.Fatalf("create plan changed: %s", plan.Raw)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(objectType, nil)}
		if d := KeepUnchangedPlan(ctx, tfsdk.State{Schema: s, Raw: state}, &plan); d.HasError() {
			t.Fatal(d)
		}
		if !plan.Raw.IsNull() {
			t.Fatalf("destroy plan changed: %s", plan.Raw)
		}
	})
}
