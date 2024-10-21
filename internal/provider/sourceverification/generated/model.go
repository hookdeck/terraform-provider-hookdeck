// Code generated by cmd/codegen. DO NOT EDIT.
// This file is automatically generated and any changes will be overwritten.
// To regenerate this file, run `go generate`.

package generated

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

type SourceVerification struct {
	JSON types.String `tfsdk:"json"`

	// Providers
	Adyen           *adyenSourceVerification           `tfsdk:"adyen"`
	Akeneo          *akeneoSourceVerification          `tfsdk:"akeneo"`
	ApiKey          *apiKeySourceVerification          `tfsdk:"api_key"`
	AwsSns          *awsSnsSourceVerification          `tfsdk:"aws_sns"`
	BasicAuth       *basicAuthSourceVerification       `tfsdk:"basic_auth"`
	Bondsmith       *bondsmithSourceVerification       `tfsdk:"bondsmith"`
	Cloudsignal     *cloudsignalSourceVerification     `tfsdk:"cloudsignal"`
	Commercelayer   *commercelayerSourceVerification   `tfsdk:"commercelayer"`
	Courier         *courierSourceVerification         `tfsdk:"courier"`
	Discord         *discordSourceVerification         `tfsdk:"discord"`
	Ebay            *ebaySourceVerification            `tfsdk:"ebay"`
	Enode           *enodeSourceVerification           `tfsdk:"enode"`
	Favro           *favroSourceVerification           `tfsdk:"favro"`
	Fiserv          *fiservSourceVerification          `tfsdk:"fiserv"`
	Frontapp        *frontappSourceVerification        `tfsdk:"frontapp"`
	Github          *githubSourceVerification          `tfsdk:"github"`
	Gitlab          *gitlabSourceVerification          `tfsdk:"gitlab"`
	Hmac            *hmacSourceVerification            `tfsdk:"hmac"`
	Hubspot         *hubspotSourceVerification         `tfsdk:"hubspot"`
	Linear          *linearSourceVerification          `tfsdk:"linear"`
	Mailchimp       *mailchimpSourceVerification       `tfsdk:"mailchimp"`
	Mailgun         *mailgunSourceVerification         `tfsdk:"mailgun"`
	Nmi             *nmiSourceVerification             `tfsdk:"nmi"`
	Orb             *orbSourceVerification             `tfsdk:"orb"`
	Oura            *ouraSourceVerification            `tfsdk:"oura"`
	Paddle          *paddleSourceVerification          `tfsdk:"paddle"`
	Paypal          *paypalSourceVerification          `tfsdk:"paypal"`
	Persona         *personaSourceVerification         `tfsdk:"persona"`
	Pipedrive       *pipedriveSourceVerification       `tfsdk:"pipedrive"`
	Postmark        *postmarkSourceVerification        `tfsdk:"postmark"`
	PropertyFinder  *propertyFinderSourceVerification  `tfsdk:"property_finder"`
	Pylon           *pylonSourceVerification           `tfsdk:"pylon"`
	Razorpay        *razorpaySourceVerification        `tfsdk:"razorpay"`
	Recharge        *rechargeSourceVerification        `tfsdk:"recharge"`
	Repay           *repaySourceVerification           `tfsdk:"repay"`
	Sanity          *sanitySourceVerification          `tfsdk:"sanity"`
	Sendgrid        *sendgridSourceVerification        `tfsdk:"sendgrid"`
	Shopify         *shopifySourceVerification         `tfsdk:"shopify"`
	Shopline        *shoplineSourceVerification        `tfsdk:"shopline"`
	Slack           *slackSourceVerification           `tfsdk:"slack"`
	Solidgate       *solidgateSourceVerification       `tfsdk:"solidgate"`
	Square          *squareSourceVerification          `tfsdk:"square"`
	Stripe          *stripeSourceVerification          `tfsdk:"stripe"`
	Svix            *svixSourceVerification            `tfsdk:"svix"`
	Synctera        *syncteraSourceVerification        `tfsdk:"synctera"`
	Tebex           *tebexSourceVerification           `tfsdk:"tebex"`
	Telnyx          *telnyxSourceVerification          `tfsdk:"telnyx"`
	ThreeDEye       *threeDEyeSourceVerification       `tfsdk:"three_d_eye"`
	Tokenio         *tokenioSourceVerification         `tfsdk:"tokenio"`
	Trello          *trelloSourceVerification          `tfsdk:"trello"`
	Twilio          *twilioSourceVerification          `tfsdk:"twilio"`
	Twitch          *twitchSourceVerification          `tfsdk:"twitch"`
	Twitter         *twitterSourceVerification         `tfsdk:"twitter"`
	Typeform        *typeformSourceVerification        `tfsdk:"typeform"`
	Vercel          *vercelSourceVerification          `tfsdk:"vercel"`
	VercelLogDrains *vercelLogDrainsSourceVerification `tfsdk:"vercel_log_drains"`
	Wix             *wixSourceVerification             `tfsdk:"wix"`
	Woocommerce     *woocommerceSourceVerification     `tfsdk:"woocommerce"`
	Workos          *workosSourceVerification          `tfsdk:"workos"`
	Xero            *xeroSourceVerification            `tfsdk:"xero"`
	Zoom            *zoomSourceVerification            `tfsdk:"zoom"`
}

type sourceVerificationProvider interface {
	getSchemaName() string
	getSchemaValue() schema.SingleNestedAttribute
	ToPayload(sourceVerification *SourceVerification) *hookdeck.VerificationConfig
}

var Providers []sourceVerificationProvider

func GetSourceVerificationSchemaAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{}

	attributes["json"] = schema.StringAttribute{
		Optional:  true,
		Sensitive: true,
	}

	for _, provider := range Providers {
		attributes[provider.getSchemaName()] = provider.getSchemaValue()
	}

	return attributes
}