package project

import "github.com/hashicorp/terraform-plugin-framework/types"

type projectResourceModel struct {
	CreatedAt      types.String `tfsdk:"created_at"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Type           types.String `tfsdk:"type"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}
