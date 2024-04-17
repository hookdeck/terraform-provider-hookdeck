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
	Cloudsignal    *cloudsignalSourceVerification    `tfsdk:"cloudsignal"`
	Commercelayer  *commercelayerSourceVerification  `tfsdk:"commercelayer"`
	Courier        *courierSourceVerification        `tfsdk:"courier"`
	Ebay           *ebaySourceVerification           `tfsdk:"ebay"`
	Enode          *enodeSourceVerification          `tfsdk:"enode"`
	Favro          *favroSourceVerification          `tfsdk:"favro"`
	FrontApp       *frontAppSourceVerification       `tfsdk:"frontapp"`
	GitHub         *githubSourceVerification         `tfsdk:"github"`
	GitLab         *gitlabSourceVerification         `tfsdk:"gitlab"`
	Linear         *linearSourceVerification         `tfsdk:"linear"`
	Mailgun        *mailgunSourceVerification        `tfsdk:"mailgun"`
	Nmi            *nmiSourceVerification            `tfsdk:"nmi"`
	Orb            *orbSourceVerification            `tfsdk:"orb"`
	Oura           *ouraSourceVerification           `tfsdk:"oura"`
	Persona        *personaSourceVerification        `tfsdk:"persona"`
	Pipedrive      *pipedriveSourceVerification      `tfsdk:"pipedrive"`
	Postmark       *postmarkSourceVerification       `tfsdk:"postmark"`
	PropertyFinder *propertyFinderSourceVerification `tfsdk:"property_finder"`
	Pylon          *pylonSourceVerification          `tfsdk:"pylon"`
	Recharge       *rechargeSourceVerification       `tfsdk:"recharge"`
	Repay          *repaySourceVerification          `tfsdk:"repay"`
	Sanity         *sanitySourceVerification         `tfsdk:"sanity"`
	SendGrid       *sendgridSourceVerification       `tfsdk:"sendgrid"`
	Shopify        *shopifySourceVerification        `tfsdk:"shopify"`
	Shopline       *shoplineSourceVerification       `tfsdk:"shopline"`
	Solidgate      *solidgateSourceVerification      `tfsdk:"solidgate"`
	Square         *squareSourceVerification         `tfsdk:"square"`
	Stripe         *stripeSourceVerification         `tfsdk:"stripe"`
	Svix           *svixSourceVerification           `tfsdk:"svix"`
	Synctera       *syncteraSourceVerification       `tfsdk:"synctera"`
	Telnyx         *telnyxSourceVerification         `tfsdk:"telnyx"`
	ThreeDEye      *threeDEyeSourceVerification      `tfsdk:"threedeye"`
	TokenIo        *tokenIoSourceVerification        `tfsdk:"tokenio"`
	Trello         *trelloSourceVerification         `tfsdk:"trello"`
	Twitch         *twitchSourceVerification         `tfsdk:"twitch"`
	Twitter        *twitterSourceVerification        `tfsdk:"twitter"`
	Typeform       *typeformSourceVerification       `tfsdk:"typeform"`
	Wix            *wixSourceVerification            `tfsdk:"wix"`
	WooCommerce    *woocommerceSourceVerification    `tfsdk:"woocommerce"`
	WorkOS         *workOSSourceVerification         `tfsdk:"workos"`
	Xero           *xeroSourceVerification           `tfsdk:"xero"`
	Zoom           *zoomSourceVerification           `tfsdk:"zoom"`
}
