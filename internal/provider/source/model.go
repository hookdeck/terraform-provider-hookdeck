package source

import "github.com/hashicorp/terraform-plugin-framework/types"

type sourceResourceModel struct {
	AllowedHTTPMethods []types.String        `tfsdk:"allowed_http_methods"`
	ArchivedAt         types.String          `tfsdk:"archived_at"`
	CreatedAt          types.String          `tfsdk:"created_at"`
	CustomResponse     *sourceCustomResponse `tfsdk:"custom_response"`
	Description        types.String          `tfsdk:"description"`
	ID                 types.String          `tfsdk:"id"`
	Name               types.String          `tfsdk:"name"`
	TeamID             types.String          `tfsdk:"team_id"`
	UpdatedAt          types.String          `tfsdk:"updated_at"`
	URL                types.String          `tfsdk:"url"`
}

type sourceCustomResponse struct {
	Body        types.String `tfsdk:"body"`
	ContentType types.String `tfsdk:"content_type"`
}
