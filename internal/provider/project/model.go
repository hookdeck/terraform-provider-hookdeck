package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type projectResourceModel struct {
	CreatedAt           types.String `tfsdk:"created_at"`
	HeadersPrefix       types.String `tfsdk:"headers_prefix"`
	ID                  types.String `tfsdk:"id"`
	MaxEventsPerSecond  types.Int64  `tfsdk:"max_events_per_second"`
	Name                types.String `tfsdk:"name"`
	NotificationMethods types.List   `tfsdk:"notification_methods"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	Type                types.String `tfsdk:"type"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}
