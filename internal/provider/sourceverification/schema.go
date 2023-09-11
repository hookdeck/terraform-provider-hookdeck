package sourceverification

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
					"commercelayer":   commercelayerConfigSchema(),
					"github":          githubConfigSchema(),
					"gitlab":          gitlabConfigSchema(),
					"mailgun":         mailgunConfigSchema(),
					"oura":            ouraConfigSchema(),
					"pipedrive":       pipedriveConfigSchema(),
					"postmark":        postmarkConfigSchema(),
					"property_finder": propertyFinderConfigSchema(),
					"recharge":        rechargeConfigSchema(),
					"sendgrid":        sendgridConfigSchema(),
					"shopify":         shopifyConfigSchema(),
					"stripe":          stripeConfigSchema(),
					"svix":            svixConfigSchema(),
					"synctera":        syncteraConfigSchema(),
					"threedeye":       threeDEyeConfigSchema(),
					"twitter":         twitterConfigSchema(),
					"typeform":        typeformConfigSchema(),
					"woocommerce":     woocommerceConfigSchema(),
					"workos":          workOSConfigSchema(),
					"xero":            xeroConfigSchema(),
					"zoom":            zoomConfigSchema(),
				},
				Description: "The verification configs for the specific verification type",
			},
		},
	}
}
