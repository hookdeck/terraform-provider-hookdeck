package sourceverification

import (
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
)

func (m *sourceVerificationResourceModel) Refresh(verification *hookdeck.VerificationConfig) {
}

func (m *sourceVerificationResourceModel) ToCreatePayload() *hookdeck.SourceUpdateRequest {
	return m.ToUpdatePayload()
}

func (m *sourceVerificationResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	var verification *hookdeck.VerificationConfig

	// generic
	if m.Verification.APIKey != nil {
		verification = m.Verification.APIKey.toPayload()
	} else if m.Verification.BasicAuth != nil {
		verification = m.Verification.BasicAuth.toPayload()
	} else if m.Verification.HMAC != nil {
		verification = m.Verification.HMAC.toPayload()
	} else if !m.Verification.JSON.IsUnknown() && !m.Verification.JSON.IsNull() {
		verification = jsonToPayload(m.Verification.JSON.ValueString())

		// providers
	} else if m.Verification.Adyen != nil {
		verification = m.Verification.Adyen.toPayload()
	} else if m.Verification.Akeneo != nil {
		verification = m.Verification.Akeneo.toPayload()
	} else if m.Verification.AWSSNS != nil {
		verification = m.Verification.AWSSNS.toPayload()
	} else if m.Verification.Bondsmith != nil {
		verification = m.Verification.Bondsmith.toPayload()
	} else if m.Verification.Cloudsignal != nil {
		verification = m.Verification.Cloudsignal.toPayload()
	} else if m.Verification.Commercelayer != nil {
		verification = m.Verification.Commercelayer.toPayload()
	} else if m.Verification.Courier != nil {
		verification = m.Verification.Courier.toPayload()
	} else if m.Verification.Discord != nil {
		verification = m.Verification.Discord.toPayload()
	} else if m.Verification.Ebay != nil {
		verification = m.Verification.Ebay.toPayload()
	} else if m.Verification.Enode != nil {
		verification = m.Verification.Enode.toPayload()
	} else if m.Verification.Favro != nil {
		verification = m.Verification.Favro.toPayload()
	} else if m.Verification.Fiserv != nil {
		verification = m.Verification.Fiserv.toPayload()
	} else if m.Verification.FrontApp != nil {
		verification = m.Verification.FrontApp.toPayload()
	} else if m.Verification.GitHub != nil {
		verification = m.Verification.GitHub.toPayload()
	} else if m.Verification.GitLab != nil {
		verification = m.Verification.GitLab.toPayload()
	} else if m.Verification.Linear != nil {
		verification = m.Verification.Linear.toPayload()
	} else if m.Verification.Mailgun != nil {
		verification = m.Verification.Mailgun.toPayload()
	} else if m.Verification.Nmi != nil {
		verification = m.Verification.Nmi.toPayload()
	} else if m.Verification.Orb != nil {
		verification = m.Verification.Orb.toPayload()
	} else if m.Verification.Oura != nil {
		verification = m.Verification.Oura.toPayload()
	} else if m.Verification.Persona != nil {
		verification = m.Verification.Persona.toPayload()
	} else if m.Verification.Pipedrive != nil {
		verification = m.Verification.Pipedrive.toPayload()
	} else if m.Verification.Postmark != nil {
		verification = m.Verification.Postmark.toPayload()
	} else if m.Verification.PropertyFinder != nil {
		verification = m.Verification.PropertyFinder.toPayload()
	} else if m.Verification.Pylon != nil {
		verification = m.Verification.Pylon.toPayload()
	} else if m.Verification.Razorpay != nil {
		verification = m.Verification.Razorpay.toPayload()
	} else if m.Verification.Recharge != nil {
		verification = m.Verification.Recharge.toPayload()
	} else if m.Verification.Repay != nil {
		verification = m.Verification.Repay.toPayload()
	} else if m.Verification.Sanity != nil {
		verification = m.Verification.Sanity.toPayload()
	} else if m.Verification.SendGrid != nil {
		verification = m.Verification.SendGrid.toPayload()
	} else if m.Verification.Shopify != nil {
		verification = m.Verification.Shopify.toPayload()
	} else if m.Verification.Shopline != nil {
		verification = m.Verification.Shopline.toPayload()
	} else if m.Verification.Slack != nil {
		verification = m.Verification.Slack.toPayload()
	} else if m.Verification.Solidgate != nil {
		verification = m.Verification.Solidgate.toPayload()
	} else if m.Verification.Square != nil {
		verification = m.Verification.Square.toPayload()
	} else if m.Verification.Stripe != nil {
		verification = m.Verification.Stripe.toPayload()
	} else if m.Verification.Svix != nil {
		verification = m.Verification.Svix.toPayload()
	} else if m.Verification.Synctera != nil {
		verification = m.Verification.Synctera.toPayload()
	} else if m.Verification.Tebex != nil {
		verification = m.Verification.Tebex.toPayload()
	} else if m.Verification.Telnyx != nil {
		verification = m.Verification.Telnyx.toPayload()
	} else if m.Verification.ThreeDEye != nil {
		verification = m.Verification.ThreeDEye.toPayload()
	} else if m.Verification.TokenIo != nil {
		verification = m.Verification.TokenIo.toPayload()
	} else if m.Verification.Trello != nil {
		verification = m.Verification.Trello.toPayload()
	} else if m.Verification.Twitch != nil {
		verification = m.Verification.Twitch.toPayload()
	} else if m.Verification.Twitter != nil {
		verification = m.Verification.Twitter.toPayload()
	} else if m.Verification.Typeform != nil {
		verification = m.Verification.Typeform.toPayload()
	} else if m.Verification.Wix != nil {
		verification = m.Verification.Wix.toPayload()
	} else if m.Verification.Vercel != nil {
		verification = m.Verification.Vercel.toPayload()
	} else if m.Verification.VercelLogDrains != nil {
		verification = m.Verification.VercelLogDrains.toPayload()
	} else if m.Verification.WooCommerce != nil {
		verification = m.Verification.WooCommerce.toPayload()
	} else if m.Verification.WorkOS != nil {
		verification = m.Verification.WorkOS.toPayload()
	} else if m.Verification.Xero != nil {
		verification = m.Verification.Xero.toPayload()
	} else if m.Verification.Zoom != nil {
		verification = m.Verification.Zoom.toPayload()
	} else {
		return &hookdeck.SourceUpdateRequest{
			Verification: hookdeck.Null[hookdeck.VerificationConfig](),
		}
	}

	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Optional(*verification),
	}
}

func (m *sourceVerificationResourceModel) ToDeletePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}
}
