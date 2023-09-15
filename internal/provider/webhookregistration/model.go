package webhookregistration

import "github.com/hashicorp/terraform-plugin-framework/types"

type webhookRegistrationResourceModel struct {
	Register   *register   `tfsdk:"register"`
	Unregister *unregister `tfsdk:"unregister"`
}

type register struct {
	Request  *request     `tfsdk:"request"`
	Response types.String `tfsdk:"response"`
}

type unregister struct {
	Request *request `tfsdk:"request"`
}

type request struct {
	Body    types.String `tfsdk:"body"`
	Headers types.String `tfsdk:"headers"`
	Method  types.String `tfsdk:"method"`
	URL     types.String `tfsdk:"url"`
}
