package shared

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestNaming(t *testing.T) {
	current := Naming{Suffix: "_source"}
	legacy := Naming{Suffix: "_source", Legacy: true}

	if got := current.TypeName("hookdeck"); got != "hookdeck_gateway_source" {
		t.Errorf("TypeName = %q", got)
	}
	if got := legacy.TypeName("hookdeck"); got != "hookdeck_source" {
		t.Errorf("legacy TypeName = %q", got)
	}
	if got := current.LegacyTypeName(); got != "hookdeck_source" {
		t.Errorf("LegacyTypeName = %q", got)
	}

	if current.ResourceDeprecation() != "" || current.DataSourceDeprecation() != "" {
		t.Error("the v3 name must not be deprecated")
	}
	if got := current.Description("Source Resource"); got != "Source Resource" {
		t.Errorf("Description = %q", got)
	}

	for name, msg := range map[string]string{
		"resource":    legacy.ResourceDeprecation(),
		"data source": legacy.DataSourceDeprecation(),
		"description": legacy.Description("Source Resource"),
	} {
		if !strings.Contains(msg, "hookdeck_source") || !strings.Contains(msg, "hookdeck_gateway_source") || !strings.Contains(msg, "v4") {
			t.Errorf("%s deprecation %q should name both types and v4", name, msg)
		}
	}
	if !strings.Contains(legacy.ResourceDeprecation(), "moved block") {
		t.Error("the resource deprecation should point at a moved block")
	}
	if strings.Contains(legacy.DataSourceDeprecation(), "moved block") {
		t.Error("data sources cannot be moved; the deprecation must not say so")
	}
}

func TestRenamedStateMover(t *testing.T) {
	schema := resourceschema.Schema{
		Version: 1,
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{Computed: true},
		},
	}
	state := tfsdk.State{
		Schema: schema,
		Raw: tftypes.NewValue(schema.Type().TerraformType(t.Context()), map[string]tftypes.Value{
			"id": tftypes.NewValue(tftypes.String, "src_1"),
		}),
	}

	cases := []struct {
		name      string
		req       resource.MoveStateRequest
		wantMoved bool
		wantErr   string
	}{
		{"terraform registry", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "registry.terraform.io/hookdeck/hookdeck", SourceSchemaVersion: 1}, true, ""},
		{"opentofu registry", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "registry.opentofu.org/hookdeck/hookdeck", SourceSchemaVersion: 1}, true, ""},
		{"mirror", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "terraform.example.com/acme/hookdeck", SourceSchemaVersion: 1}, true, ""},
		{"bare address from OpenTofu 1.10 to 1.12.3", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "hookdeck", SourceSchemaVersion: 1}, true, ""},
		{"another type", resource.MoveStateRequest{SourceTypeName: "hookdeck_destination", SourceProviderAddress: "registry.terraform.io/hookdeck/hookdeck", SourceSchemaVersion: 1}, false, ""},
		{"another provider", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "registry.terraform.io/acme/other", SourceSchemaVersion: 1}, false, ""},
		{"provider whose name ends in hookdeck", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "registry.terraform.io/acme/nothookdeck", SourceSchemaVersion: 1}, false, ""},
		{"older schema version", resource.MoveStateRequest{SourceTypeName: "hookdeck_source", SourceProviderAddress: "registry.terraform.io/hookdeck/hookdeck", SourceSchemaVersion: 0}, false, "Unsupported state version for move"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mover := RenamedStateMover("hookdeck_source", schema, schema)
			tc.req.SourceState = &state
			resp := &resource.MoveStateResponse{}

			mover.StateMover(t.Context(), tc.req, resp)

			if tc.wantErr != "" {
				if !resp.Diagnostics.HasError() || resp.Diagnostics[0].Summary() != tc.wantErr {
					t.Fatalf("diagnostics = %v, want %q", resp.Diagnostics, tc.wantErr)
				}
				return
			}
			if resp.Diagnostics.HasError() {
				t.Fatalf("diagnostics = %v", resp.Diagnostics)
			}
			moved := !resp.TargetState.Raw.IsNull()
			if moved != tc.wantMoved {
				t.Fatalf("moved = %v, want %v", moved, tc.wantMoved)
			}
			if !moved {
				return
			}
			if !resp.TargetState.Raw.Equal(state.Raw) {
				t.Error("state was not copied as-is")
			}
		})
	}
}

func TestRenamedStateMover_undecodedState(t *testing.T) {
	schema := resourceschema.Schema{Attributes: map[string]resourceschema.Attribute{
		"id": resourceschema.StringAttribute{Computed: true},
	}}
	resp := &resource.MoveStateResponse{}
	RenamedStateMover("hookdeck_source", schema, schema).StateMover(t.Context(), resource.MoveStateRequest{
		SourceTypeName:        "hookdeck_source",
		SourceProviderAddress: "registry.terraform.io/hookdeck/hookdeck",
	}, resp)
	if !resp.Diagnostics.HasError() || resp.Diagnostics[0].Summary() != "Unsupported state for move" {
		t.Fatalf("diagnostics = %v", resp.Diagnostics)
	}
}

// The v3 names have no team_id. Moving drops it, and state written before
// project_id existed gets its project from it.
func TestRenamedStateMover_teamID(t *testing.T) {
	target := resourceschema.Schema{Version: 1, Attributes: map[string]resourceschema.Attribute{
		"id":         resourceschema.StringAttribute{Computed: true},
		"project_id": ProjectIDResourceAttribute(),
	}}
	source := resourceschema.Schema{Version: 1, Attributes: Naming{Suffix: "_source", Legacy: true}.WithTeamID(map[string]resourceschema.Attribute{
		"id":         resourceschema.StringAttribute{Computed: true},
		"project_id": ProjectIDResourceAttribute(),
	})}
	if _, ok := source.Attributes["team_id"]; !ok {
		t.Fatal("the v2 schema has no team_id")
	}
	if _, ok := (Naming{Suffix: "_source"}).WithTeamID(map[string]resourceschema.Attribute{})["team_id"]; ok {
		t.Fatal("the v3 schema has team_id")
	}

	str := func(v string) tftypes.Value {
		if v == "" {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, v)
	}
	cases := []struct {
		name, projectID, teamID, want string
	}{
		{"v3 state", "tm_p", "tm_p", "tm_p"},
		{"state written before project_id existed", "", "tm_t", "tm_t"},
		{"project_id wins", "tm_p", "tm_t", "tm_p"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := tfsdk.State{Schema: source, Raw: tftypes.NewValue(source.Type().TerraformType(t.Context()), map[string]tftypes.Value{
				"id": str("src_1"), "project_id": str(tc.projectID), "team_id": str(tc.teamID),
			})}
			resp := &resource.MoveStateResponse{TargetState: tfsdk.State{Schema: target}}
			RenamedStateMover("hookdeck_source", source, target).StateMover(t.Context(), resource.MoveStateRequest{
				SourceTypeName:        "hookdeck_source",
				SourceProviderAddress: "registry.terraform.io/hookdeck/hookdeck",
				SourceSchemaVersion:   1,
				SourceState:           &state,
			}, resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			want := tftypes.NewValue(target.Type().TerraformType(t.Context()), map[string]tftypes.Value{
				"id": str("src_1"), "project_id": str(tc.want),
			})
			if !resp.TargetState.Raw.Equal(want) {
				t.Errorf("moved state = %v, want %v", resp.TargetState.Raw, want)
			}
		})
	}
}
