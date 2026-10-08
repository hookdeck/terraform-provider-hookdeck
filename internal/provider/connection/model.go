package connection

import (
	"context"

	"terraform-provider-hookdeck/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type connectionResourceModel struct {
	ProjectID     types.String `tfsdk:"project_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	Description   types.String `tfsdk:"description"`
	DestinationID types.String `tfsdk:"destination_id"`
	DisabledAt    types.String `tfsdk:"disabled_at"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PausedAt      types.String `tfsdk:"paused_at"`
	Rules         []rule       `tfsdk:"rules"`
	SourceID      types.String `tfsdk:"source_id"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type rule struct {
	DeduplicateRule *deduplicateRule `tfsdk:"deduplicate_rule"`
	DelayRule       *delayRule       `tfsdk:"delay_rule"`
	FilterRule      *filterRule      `tfsdk:"filter_rule"`
	RetryRule       *retryRule       `tfsdk:"retry_rule"`
	TransformRule   *transformRule   `tfsdk:"transform_rule"`
}

type delayRule struct {
	Delay types.Int64 `tfsdk:"delay"`
}

type filterRule struct {
	Body    *filterRuleProperty `tfsdk:"body"`
	Headers *filterRuleProperty `tfsdk:"headers"`
	Path    *filterRuleProperty `tfsdk:"path"`
	Query   *filterRuleProperty `tfsdk:"query"`
}

type filterRuleProperty struct {
	Boolean types.Bool           `tfsdk:"boolean"`
	JSON    jsontypes.Normalized `tfsdk:"json"`
	Number  types.Number         `tfsdk:"number"`
	String  types.String         `tfsdk:"string"`
}

type retryRule struct {
	Count               types.Int64  `tfsdk:"count"`
	Interval            types.Int64  `tfsdk:"interval"`
	Strategy            types.String `tfsdk:"strategy"`
	ResponseStatusCodes types.List   `tfsdk:"response_status_codes"`
}

type transformRule struct {
	TransformationID types.String `tfsdk:"transformation_id"`
}

type deduplicateRule struct {
	Window        types.Int64 `tfsdk:"window"`
	IncludeFields types.List  `tfsdk:"include_fields"`
	ExcludeFields types.List  `tfsdk:"exclude_fields"`
}

// V0 model structs for state upgrade compatibility
// These match the V0 schema which:
// - Did NOT have deduplicate_rule
// - Did NOT have response_status_codes in retry_rule
// - Used plain types.String for JSON in filter rule (not jsontypes.Normalized)

type connectionResourceModelV0 struct {
	CreatedAt     types.String `tfsdk:"created_at"`
	Description   types.String `tfsdk:"description"`
	DestinationID types.String `tfsdk:"destination_id"`
	DisabledAt    types.String `tfsdk:"disabled_at"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PausedAt      types.String `tfsdk:"paused_at"`
	Rules         []ruleV0     `tfsdk:"rules"`
	SourceID      types.String `tfsdk:"source_id"`
	TeamID        types.String `tfsdk:"team_id"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type ruleV0 struct {
	DelayRule     *delayRule     `tfsdk:"delay_rule"`
	FilterRule    *filterRuleV0  `tfsdk:"filter_rule"`
	RetryRule     *retryRuleV0   `tfsdk:"retry_rule"`
	TransformRule *transformRule `tfsdk:"transform_rule"`
}

type filterRuleV0 struct {
	Body    *filterRulePropertyV0 `tfsdk:"body"`
	Headers *filterRulePropertyV0 `tfsdk:"headers"`
	Path    *filterRulePropertyV0 `tfsdk:"path"`
	Query   *filterRulePropertyV0 `tfsdk:"query"`
}

type filterRulePropertyV0 struct {
	Boolean types.Bool   `tfsdk:"boolean"`
	JSON    types.String `tfsdk:"json"` // V0 used plain string, not jsontypes.Normalized
	Number  types.Number `tfsdk:"number"`
	String  types.String `tfsdk:"string"`
}

type retryRuleV0 struct {
	Count    types.Int64  `tfsdk:"count"`
	Interval types.Int64  `tfsdk:"interval"`
	Strategy types.String `tfsdk:"strategy"`
	// Note: V0 did NOT have response_status_codes
}

// legacyConnectionResourceModel is the model of the v2 name, which also has team_id.
type legacyConnectionResourceModel struct {
	connectionResourceModel
	TeamID types.String `tfsdk:"team_id"`
}

// getModel reads the model from a plan, state or configuration. On the v2
// name it also returns team_id, which records the project in state written
// before project_id existed.
func getModel(ctx context.Context, from shared.ModelSource, legacy bool) (*connectionResourceModel, types.String, diag.Diagnostics) {
	if legacy {
		var m legacyConnectionResourceModel
		diags := from.Get(ctx, &m)
		return &m.connectionResourceModel, m.TeamID, diags
	}
	var m connectionResourceModel
	diags := from.Get(ctx, &m)
	return &m, types.StringNull(), diags
}

// setModel writes the model to state, with team_id on the v2 name.
func setModel(ctx context.Context, to shared.ModelTarget, legacy bool, m *connectionResourceModel) diag.Diagnostics {
	if legacy {
		return to.Set(ctx, &legacyConnectionResourceModel{connectionResourceModel: *m, TeamID: m.ProjectID})
	}
	return to.Set(ctx, m)
}
