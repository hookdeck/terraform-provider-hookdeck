package sourceverification

import "github.com/hashicorp/terraform-plugin-framework/types"

type sourceVerificationResourceModel struct {
	SourceID     types.String        `tfsdk:"source_id"`
	Verification *sourceVerification `tfsdk:"verification"`
}

type sourceVerification struct {
	// generic
	APIKey    *apiKeySourceVerification    `tfsdk:"api_key"`
	BasicAuth *basicAuthSourceVerification `tfsdk:"basic_auth"`
	HMAC      *hmacSourceVerification      `tfsdk:"hmac"`
	// providers
	Adyen          *adyenSourceVerification          `tfsdk:"adyen"`
	Akeneo         *akeneoSourceVerification         `tfsdk:"akeneo"`
	AWSSNS         *awsSNSSourceVerification         `tfsdk:"aws_sns"`
	Commercelayer  *commercelayerSourceVerification  `tfsdk:"commercelayer"`
	GitHub         *githubSourceVerification         `tfsdk:"github"`
	GitLab         *gitlabSourceVerification         `tfsdk:"gitlab"`
	Mailgun        *mailgunSourceVerification        `tfsdk:"mailgun"`
	Oura           *ouraSourceVerification           `tfsdk:"oura"`
	Pipedrive      *pipedriveSourceVerification      `tfsdk:"pipedrive"`
	Postmark       *postmarkSourceVerification       `tfsdk:"postmark"`
	PropertyFinder *propertyFinderSourceVerification `tfsdk:"property_finder"`
	Recharge       *rechargeSourceVerification       `tfsdk:"recharge"`
	SendGrid       *sendgridSourceVerification       `tfsdk:"sendgrid"`
	Shopify        *shopifySourceVerification        `tfsdk:"shopify"`
	Stripe         *stripeSourceVerification         `tfsdk:"stripe"`
	Svix           *svixSourceVerification           `tfsdk:"svix"`
	Synctera       *syncteraSourceVerification       `tfsdk:"synctera"`
	ThreeDEye      *threeDEyeSourceVerification      `tfsdk:"threedeye"`
	Twitter        *twitterSourceVerification        `tfsdk:"twitter"`
	Typeform       *typeformSourceVerification       `tfsdk:"typeform"`
	WooCommerce    *woocommerceSourceVerification    `tfsdk:"woocommerce"`
	WorkOS         *workOSSourceVerification         `tfsdk:"workos"`
	Xero           *xeroSourceVerification           `tfsdk:"xero"`
	Zoom           *zoomSourceVerification           `tfsdk:"zoom"`
}
