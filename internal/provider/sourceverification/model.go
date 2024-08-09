package sourceverification

import "github.com/hashicorp/terraform-plugin-framework/types"

type sourceVerificationResourceModel struct {
	SourceID     types.String        `tfsdk:"source_id"`
	Verification *sourceVerification `tfsdk:"verification"`
}

type sourceVerification struct {
	// generic
	ApiKey *apiKeySourceVerification `tfsdk:"api_key"`
	Stripe *stripeSourceVerification `tfsdk:"stripe"`
}
