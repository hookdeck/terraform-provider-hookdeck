package sourceverification

import (
	"context"
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *sourceVerificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Transformation Resource",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ID of the source",
			},
			"verification": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					// generic
					"api_key":    apiKeyConfigSchema(),
					"basic_auth": basicAuthConfigSchema(),
					"hmac":       hmacConfigSchema(),
					// providers
					"adyen":           adyenConfigSchema(),
					"akeneo":          akeneoConfigSchema(),
					"aws_sns":         awsSNSConfigSchema(),
					"cloudsignal":     cloudsignalConfigSchema(),
					"commercelayer":   commercelayerConfigSchema(),
					"courier":         courierConfigSchema(),
					"ebay":            ebayConfigSchema(),
					"enode":           enodeConfigSchema(),
					"favro":           favroConfigSchema(),
					"frontapp":        frontAppConfigSchema(),
					"github":          githubConfigSchema(),
					"gitlab":          gitlabConfigSchema(),
					"linear":          linearConfigSchema(),
					"mailgun":         mailgunConfigSchema(),
					"nmi":             nmiConfigSchema(),
					"orb":             orbConfigSchema(),
					"oura":            ouraConfigSchema(),
					"persona":         personaConfigSchema(),
					"pipedrive":       pipedriveConfigSchema(),
					"postmark":        postmarkConfigSchema(),
					"property_finder": propertyFinderConfigSchema(),
					"pylon":           pylonConfigSchema(),
					"recharge":        rechargeConfigSchema(),
					"repay":           repayConfigSchema(),
					"sanity":          sanityConfigSchema(),
					"sendgrid":        sendgridConfigSchema(),
					"shopify":         shopifyConfigSchema(),
					"shopline":        shoplineConfigSchema(),
					"solidgate":       solidgateConfigSchema(),
					"square":          squareConfigSchema(),
					"stripe":          stripeConfigSchema(),
					"svix":            svixConfigSchema(),
					"synctera":        syncteraConfigSchema(),
					"telnyx":          telnyxConfigSchema(),
					"threedeye":       threeDEyeConfigSchema(),
					"tokenio":         tokenIoConfigSchema(),
					"trello":          trelloConfigSchema(),
					"twitch":          twitchConfigSchema(),
					"twitter":         twitterConfigSchema(),
					"typeform":        typeformConfigSchema(),
					"wix":             wixConfigSchema(),
					"woocommerce":     woocommerceConfigSchema(),
					"workos":          workOSConfigSchema(),
					"xero":            xeroConfigSchema(),
					"zoom":            zoomConfigSchema(),
				},
				Validators: []validator.Object{
					validators.ExactlyOneChild(),
				},
				Description: "The verification configs for the specific verification type",
			},
		},
	}
}
