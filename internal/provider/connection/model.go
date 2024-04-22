package connection

import "github.com/hashicorp/terraform-plugin-framework/types"

type connectionResourceModel struct {
	CreatedAt     types.String `tfsdk:"created_at"`
	Description   types.String `tfsdk:"description"`
	DestinationID types.String `tfsdk:"destination_id"`
	DisabledAt    types.String `tfsdk:"disabled_at"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PausedAt      types.String `tfsdk:"paused_at"`
	Rules         []rule       `tfsdk:"rules"`
	SourceID      types.String `tfsdk:"source_id"`
	TeamID        types.String `tfsdk:"team_id"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type rule struct {
	DelayRule     *delayRule     `tfsdk:"delay_rule"`
	FilterRule    *filterRule    `tfsdk:"filter_rule"`
	RetryRule     *retryRule     `tfsdk:"retry_rule"`
	TransformRule *transformRule `tfsdk:"transform_rule"`
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
	Boolean types.Bool   `tfsdk:"boolean"`
	JSON    types.String `tfsdk:"json"`
	Number  types.Number `tfsdk:"number"`
	String  types.String `tfsdk:"string"`
}

type retryRule struct {
	Count    types.Int64  `tfsdk:"count"`
	Interval types.Int64  `tfsdk:"interval"`
	Strategy types.String `tfsdk:"strategy"`
}

type transformRule struct {
	TransformationID types.String `tfsdk:"transformation_id"`
}
