// This file was auto-generated by Fern from our API Definition.

package api

import (
	json "encoding/json"
	fmt "fmt"
	core "github.com/hookdeck/hookdeck-go-sdk/core"
	internal "github.com/hookdeck/hookdeck-go-sdk/internal"
	time "time"
)

type SourceCreateRequest struct {
	// A unique name for the source
	Name string `json:"name" url:"-"`
	// Type of the source
	Type *core.Optional[SourceCreateRequestType] `json:"type,omitempty" url:"-"`
	// Description for the source
	Description *core.Optional[string]           `json:"description,omitempty" url:"-"`
	Config      *core.Optional[SourceTypeConfig] `json:"config,omitempty" url:"-"`
}

type SourceListRequest struct {
	Id         []*string                 `json:"-" url:"id,omitempty"`
	Name       *string                   `json:"-" url:"name,omitempty"`
	Disabled   *bool                     `json:"-" url:"disabled,omitempty"`
	DisabledAt *time.Time                `json:"-" url:"disabled_at,omitempty"`
	OrderBy    *SourceListRequestOrderBy `json:"-" url:"order_by,omitempty"`
	Dir        *SourceListRequestDir     `json:"-" url:"dir,omitempty"`
	Limit      *int                      `json:"-" url:"limit,omitempty"`
	Next       *string                   `json:"-" url:"next,omitempty"`
	Prev       *string                   `json:"-" url:"prev,omitempty"`
}

type SourceRetrieveRequest struct {
	Include *string `json:"-" url:"include,omitempty"`
}

type SourcePaginatedResult struct {
	Pagination *SeekPagination `json:"pagination,omitempty" url:"pagination,omitempty"`
	Count      *int            `json:"count,omitempty" url:"count,omitempty"`
	Models     []*Source       `json:"models,omitempty" url:"models,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (s *SourcePaginatedResult) GetPagination() *SeekPagination {
	if s == nil {
		return nil
	}
	return s.Pagination
}

func (s *SourcePaginatedResult) GetCount() *int {
	if s == nil {
		return nil
	}
	return s.Count
}

func (s *SourcePaginatedResult) GetModels() []*Source {
	if s == nil {
		return nil
	}
	return s.Models
}

func (s *SourcePaginatedResult) GetExtraProperties() map[string]interface{} {
	return s.extraProperties
}

func (s *SourcePaginatedResult) UnmarshalJSON(data []byte) error {
	type unmarshaler SourcePaginatedResult
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = SourcePaginatedResult(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *s)
	if err != nil {
		return err
	}
	s.extraProperties = extraProperties
	s.rawJSON = json.RawMessage(data)
	return nil
}

func (s *SourcePaginatedResult) String() string {
	if len(s.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(s.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(s); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", s)
}

// Type of the source
type SourceCreateRequestType string

const (
	SourceCreateRequestTypeWebhook         SourceCreateRequestType = "WEBHOOK"
	SourceCreateRequestTypeHttp            SourceCreateRequestType = "HTTP"
	SourceCreateRequestTypeManaged         SourceCreateRequestType = "MANAGED"
	SourceCreateRequestTypeSanity          SourceCreateRequestType = "SANITY"
	SourceCreateRequestTypeBridge          SourceCreateRequestType = "BRIDGE"
	SourceCreateRequestTypeCloudsignal     SourceCreateRequestType = "CLOUDSIGNAL"
	SourceCreateRequestTypeCourier         SourceCreateRequestType = "COURIER"
	SourceCreateRequestTypeFrontapp        SourceCreateRequestType = "FRONTAPP"
	SourceCreateRequestTypeZoom            SourceCreateRequestType = "ZOOM"
	SourceCreateRequestTypeTwitter         SourceCreateRequestType = "TWITTER"
	SourceCreateRequestTypeRecharge        SourceCreateRequestType = "RECHARGE"
	SourceCreateRequestTypeStripe          SourceCreateRequestType = "STRIPE"
	SourceCreateRequestTypePropertyFinder  SourceCreateRequestType = "PROPERTY-FINDER"
	SourceCreateRequestTypeShopify         SourceCreateRequestType = "SHOPIFY"
	SourceCreateRequestTypeTwilio          SourceCreateRequestType = "TWILIO"
	SourceCreateRequestTypeGithub          SourceCreateRequestType = "GITHUB"
	SourceCreateRequestTypePostmark        SourceCreateRequestType = "POSTMARK"
	SourceCreateRequestTypeTypeform        SourceCreateRequestType = "TYPEFORM"
	SourceCreateRequestTypeXero            SourceCreateRequestType = "XERO"
	SourceCreateRequestTypeSvix            SourceCreateRequestType = "SVIX"
	SourceCreateRequestTypeAdyen           SourceCreateRequestType = "ADYEN"
	SourceCreateRequestTypeAkeneo          SourceCreateRequestType = "AKENEO"
	SourceCreateRequestTypeGitlab          SourceCreateRequestType = "GITLAB"
	SourceCreateRequestTypeWoocommerce     SourceCreateRequestType = "WOOCOMMERCE"
	SourceCreateRequestTypeOura            SourceCreateRequestType = "OURA"
	SourceCreateRequestTypeCommercelayer   SourceCreateRequestType = "COMMERCELAYER"
	SourceCreateRequestTypeHubspot         SourceCreateRequestType = "HUBSPOT"
	SourceCreateRequestTypeMailgun         SourceCreateRequestType = "MAILGUN"
	SourceCreateRequestTypePersona         SourceCreateRequestType = "PERSONA"
	SourceCreateRequestTypePipedrive       SourceCreateRequestType = "PIPEDRIVE"
	SourceCreateRequestTypeSendgrid        SourceCreateRequestType = "SENDGRID"
	SourceCreateRequestTypeWorkos          SourceCreateRequestType = "WORKOS"
	SourceCreateRequestTypeSynctera        SourceCreateRequestType = "SYNCTERA"
	SourceCreateRequestTypeAwsSns          SourceCreateRequestType = "AWS_SNS"
	SourceCreateRequestTypeThreeDEye       SourceCreateRequestType = "THREE_D_EYE"
	SourceCreateRequestTypeTwitch          SourceCreateRequestType = "TWITCH"
	SourceCreateRequestTypeEnode           SourceCreateRequestType = "ENODE"
	SourceCreateRequestTypeFavro           SourceCreateRequestType = "FAVRO"
	SourceCreateRequestTypeLinear          SourceCreateRequestType = "LINEAR"
	SourceCreateRequestTypeShopline        SourceCreateRequestType = "SHOPLINE"
	SourceCreateRequestTypeWix             SourceCreateRequestType = "WIX"
	SourceCreateRequestTypeNmi             SourceCreateRequestType = "NMI"
	SourceCreateRequestTypeOrb             SourceCreateRequestType = "ORB"
	SourceCreateRequestTypePylon           SourceCreateRequestType = "PYLON"
	SourceCreateRequestTypeRazorpay        SourceCreateRequestType = "RAZORPAY"
	SourceCreateRequestTypeRepay           SourceCreateRequestType = "REPAY"
	SourceCreateRequestTypeSquare          SourceCreateRequestType = "SQUARE"
	SourceCreateRequestTypeSolidgate       SourceCreateRequestType = "SOLIDGATE"
	SourceCreateRequestTypeTrello          SourceCreateRequestType = "TRELLO"
	SourceCreateRequestTypeEbay            SourceCreateRequestType = "EBAY"
	SourceCreateRequestTypeTelnyx          SourceCreateRequestType = "TELNYX"
	SourceCreateRequestTypeDiscord         SourceCreateRequestType = "DISCORD"
	SourceCreateRequestTypeTokenio         SourceCreateRequestType = "TOKENIO"
	SourceCreateRequestTypeFiserv          SourceCreateRequestType = "FISERV"
	SourceCreateRequestTypeBondsmith       SourceCreateRequestType = "BONDSMITH"
	SourceCreateRequestTypeVercelLogDrains SourceCreateRequestType = "VERCEL_LOG_DRAINS"
	SourceCreateRequestTypeVercel          SourceCreateRequestType = "VERCEL"
	SourceCreateRequestTypeTebex           SourceCreateRequestType = "TEBEX"
	SourceCreateRequestTypeSlack           SourceCreateRequestType = "SLACK"
	SourceCreateRequestTypeMailchimp       SourceCreateRequestType = "MAILCHIMP"
	SourceCreateRequestTypePaddle          SourceCreateRequestType = "PADDLE"
	SourceCreateRequestTypePaypal          SourceCreateRequestType = "PAYPAL"
	SourceCreateRequestTypeTreezor         SourceCreateRequestType = "TREEZOR"
	SourceCreateRequestTypePraxis          SourceCreateRequestType = "PRAXIS"
	SourceCreateRequestTypeCustomerio      SourceCreateRequestType = "CUSTOMERIO"
	SourceCreateRequestTypeFacebook        SourceCreateRequestType = "FACEBOOK"
	SourceCreateRequestTypeWhatsapp        SourceCreateRequestType = "WHATSAPP"
	SourceCreateRequestTypeReplicate       SourceCreateRequestType = "REPLICATE"
	SourceCreateRequestTypeTiktok          SourceCreateRequestType = "TIKTOK"
	SourceCreateRequestTypeAirwallex       SourceCreateRequestType = "AIRWALLEX"
	SourceCreateRequestTypeZendesk         SourceCreateRequestType = "ZENDESK"
	SourceCreateRequestTypeUpollo          SourceCreateRequestType = "UPOLLO"
	SourceCreateRequestTypeLinkedin        SourceCreateRequestType = "LINKEDIN"
)

func NewSourceCreateRequestTypeFromString(s string) (SourceCreateRequestType, error) {
	switch s {
	case "WEBHOOK":
		return SourceCreateRequestTypeWebhook, nil
	case "HTTP":
		return SourceCreateRequestTypeHttp, nil
	case "MANAGED":
		return SourceCreateRequestTypeManaged, nil
	case "SANITY":
		return SourceCreateRequestTypeSanity, nil
	case "BRIDGE":
		return SourceCreateRequestTypeBridge, nil
	case "CLOUDSIGNAL":
		return SourceCreateRequestTypeCloudsignal, nil
	case "COURIER":
		return SourceCreateRequestTypeCourier, nil
	case "FRONTAPP":
		return SourceCreateRequestTypeFrontapp, nil
	case "ZOOM":
		return SourceCreateRequestTypeZoom, nil
	case "TWITTER":
		return SourceCreateRequestTypeTwitter, nil
	case "RECHARGE":
		return SourceCreateRequestTypeRecharge, nil
	case "STRIPE":
		return SourceCreateRequestTypeStripe, nil
	case "PROPERTY-FINDER":
		return SourceCreateRequestTypePropertyFinder, nil
	case "SHOPIFY":
		return SourceCreateRequestTypeShopify, nil
	case "TWILIO":
		return SourceCreateRequestTypeTwilio, nil
	case "GITHUB":
		return SourceCreateRequestTypeGithub, nil
	case "POSTMARK":
		return SourceCreateRequestTypePostmark, nil
	case "TYPEFORM":
		return SourceCreateRequestTypeTypeform, nil
	case "XERO":
		return SourceCreateRequestTypeXero, nil
	case "SVIX":
		return SourceCreateRequestTypeSvix, nil
	case "ADYEN":
		return SourceCreateRequestTypeAdyen, nil
	case "AKENEO":
		return SourceCreateRequestTypeAkeneo, nil
	case "GITLAB":
		return SourceCreateRequestTypeGitlab, nil
	case "WOOCOMMERCE":
		return SourceCreateRequestTypeWoocommerce, nil
	case "OURA":
		return SourceCreateRequestTypeOura, nil
	case "COMMERCELAYER":
		return SourceCreateRequestTypeCommercelayer, nil
	case "HUBSPOT":
		return SourceCreateRequestTypeHubspot, nil
	case "MAILGUN":
		return SourceCreateRequestTypeMailgun, nil
	case "PERSONA":
		return SourceCreateRequestTypePersona, nil
	case "PIPEDRIVE":
		return SourceCreateRequestTypePipedrive, nil
	case "SENDGRID":
		return SourceCreateRequestTypeSendgrid, nil
	case "WORKOS":
		return SourceCreateRequestTypeWorkos, nil
	case "SYNCTERA":
		return SourceCreateRequestTypeSynctera, nil
	case "AWS_SNS":
		return SourceCreateRequestTypeAwsSns, nil
	case "THREE_D_EYE":
		return SourceCreateRequestTypeThreeDEye, nil
	case "TWITCH":
		return SourceCreateRequestTypeTwitch, nil
	case "ENODE":
		return SourceCreateRequestTypeEnode, nil
	case "FAVRO":
		return SourceCreateRequestTypeFavro, nil
	case "LINEAR":
		return SourceCreateRequestTypeLinear, nil
	case "SHOPLINE":
		return SourceCreateRequestTypeShopline, nil
	case "WIX":
		return SourceCreateRequestTypeWix, nil
	case "NMI":
		return SourceCreateRequestTypeNmi, nil
	case "ORB":
		return SourceCreateRequestTypeOrb, nil
	case "PYLON":
		return SourceCreateRequestTypePylon, nil
	case "RAZORPAY":
		return SourceCreateRequestTypeRazorpay, nil
	case "REPAY":
		return SourceCreateRequestTypeRepay, nil
	case "SQUARE":
		return SourceCreateRequestTypeSquare, nil
	case "SOLIDGATE":
		return SourceCreateRequestTypeSolidgate, nil
	case "TRELLO":
		return SourceCreateRequestTypeTrello, nil
	case "EBAY":
		return SourceCreateRequestTypeEbay, nil
	case "TELNYX":
		return SourceCreateRequestTypeTelnyx, nil
	case "DISCORD":
		return SourceCreateRequestTypeDiscord, nil
	case "TOKENIO":
		return SourceCreateRequestTypeTokenio, nil
	case "FISERV":
		return SourceCreateRequestTypeFiserv, nil
	case "BONDSMITH":
		return SourceCreateRequestTypeBondsmith, nil
	case "VERCEL_LOG_DRAINS":
		return SourceCreateRequestTypeVercelLogDrains, nil
	case "VERCEL":
		return SourceCreateRequestTypeVercel, nil
	case "TEBEX":
		return SourceCreateRequestTypeTebex, nil
	case "SLACK":
		return SourceCreateRequestTypeSlack, nil
	case "MAILCHIMP":
		return SourceCreateRequestTypeMailchimp, nil
	case "PADDLE":
		return SourceCreateRequestTypePaddle, nil
	case "PAYPAL":
		return SourceCreateRequestTypePaypal, nil
	case "TREEZOR":
		return SourceCreateRequestTypeTreezor, nil
	case "PRAXIS":
		return SourceCreateRequestTypePraxis, nil
	case "CUSTOMERIO":
		return SourceCreateRequestTypeCustomerio, nil
	case "FACEBOOK":
		return SourceCreateRequestTypeFacebook, nil
	case "WHATSAPP":
		return SourceCreateRequestTypeWhatsapp, nil
	case "REPLICATE":
		return SourceCreateRequestTypeReplicate, nil
	case "TIKTOK":
		return SourceCreateRequestTypeTiktok, nil
	case "AIRWALLEX":
		return SourceCreateRequestTypeAirwallex, nil
	case "ZENDESK":
		return SourceCreateRequestTypeZendesk, nil
	case "UPOLLO":
		return SourceCreateRequestTypeUpollo, nil
	case "LINKEDIN":
		return SourceCreateRequestTypeLinkedin, nil
	}
	var t SourceCreateRequestType
	return "", fmt.Errorf("%s is not a valid %T", s, t)
}

func (s SourceCreateRequestType) Ptr() *SourceCreateRequestType {
	return &s
}

type SourceDeleteResponse struct {
	// ID of the source
	Id string `json:"id" url:"id"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (s *SourceDeleteResponse) GetId() string {
	if s == nil {
		return ""
	}
	return s.Id
}

func (s *SourceDeleteResponse) GetExtraProperties() map[string]interface{} {
	return s.extraProperties
}

func (s *SourceDeleteResponse) UnmarshalJSON(data []byte) error {
	type unmarshaler SourceDeleteResponse
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = SourceDeleteResponse(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *s)
	if err != nil {
		return err
	}
	s.extraProperties = extraProperties
	s.rawJSON = json.RawMessage(data)
	return nil
}

func (s *SourceDeleteResponse) String() string {
	if len(s.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(s.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(s); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", s)
}

type SourceListRequestDir string

const (
	SourceListRequestDirAsc  SourceListRequestDir = "asc"
	SourceListRequestDirDesc SourceListRequestDir = "desc"
)

func NewSourceListRequestDirFromString(s string) (SourceListRequestDir, error) {
	switch s {
	case "asc":
		return SourceListRequestDirAsc, nil
	case "desc":
		return SourceListRequestDirDesc, nil
	}
	var t SourceListRequestDir
	return "", fmt.Errorf("%s is not a valid %T", s, t)
}

func (s SourceListRequestDir) Ptr() *SourceListRequestDir {
	return &s
}

type SourceListRequestOrderBy string

const (
	SourceListRequestOrderByCreatedAt SourceListRequestOrderBy = "created_at"
)

func NewSourceListRequestOrderByFromString(s string) (SourceListRequestOrderBy, error) {
	switch s {
	case "created_at":
		return SourceListRequestOrderByCreatedAt, nil
	}
	var t SourceListRequestOrderBy
	return "", fmt.Errorf("%s is not a valid %T", s, t)
}

func (s SourceListRequestOrderBy) Ptr() *SourceListRequestOrderBy {
	return &s
}

// Type of the source
type SourceUpdateRequestType string

const (
	SourceUpdateRequestTypeWebhook         SourceUpdateRequestType = "WEBHOOK"
	SourceUpdateRequestTypeHttp            SourceUpdateRequestType = "HTTP"
	SourceUpdateRequestTypeManaged         SourceUpdateRequestType = "MANAGED"
	SourceUpdateRequestTypeSanity          SourceUpdateRequestType = "SANITY"
	SourceUpdateRequestTypeBridge          SourceUpdateRequestType = "BRIDGE"
	SourceUpdateRequestTypeCloudsignal     SourceUpdateRequestType = "CLOUDSIGNAL"
	SourceUpdateRequestTypeCourier         SourceUpdateRequestType = "COURIER"
	SourceUpdateRequestTypeFrontapp        SourceUpdateRequestType = "FRONTAPP"
	SourceUpdateRequestTypeZoom            SourceUpdateRequestType = "ZOOM"
	SourceUpdateRequestTypeTwitter         SourceUpdateRequestType = "TWITTER"
	SourceUpdateRequestTypeRecharge        SourceUpdateRequestType = "RECHARGE"
	SourceUpdateRequestTypeStripe          SourceUpdateRequestType = "STRIPE"
	SourceUpdateRequestTypePropertyFinder  SourceUpdateRequestType = "PROPERTY-FINDER"
	SourceUpdateRequestTypeShopify         SourceUpdateRequestType = "SHOPIFY"
	SourceUpdateRequestTypeTwilio          SourceUpdateRequestType = "TWILIO"
	SourceUpdateRequestTypeGithub          SourceUpdateRequestType = "GITHUB"
	SourceUpdateRequestTypePostmark        SourceUpdateRequestType = "POSTMARK"
	SourceUpdateRequestTypeTypeform        SourceUpdateRequestType = "TYPEFORM"
	SourceUpdateRequestTypeXero            SourceUpdateRequestType = "XERO"
	SourceUpdateRequestTypeSvix            SourceUpdateRequestType = "SVIX"
	SourceUpdateRequestTypeAdyen           SourceUpdateRequestType = "ADYEN"
	SourceUpdateRequestTypeAkeneo          SourceUpdateRequestType = "AKENEO"
	SourceUpdateRequestTypeGitlab          SourceUpdateRequestType = "GITLAB"
	SourceUpdateRequestTypeWoocommerce     SourceUpdateRequestType = "WOOCOMMERCE"
	SourceUpdateRequestTypeOura            SourceUpdateRequestType = "OURA"
	SourceUpdateRequestTypeCommercelayer   SourceUpdateRequestType = "COMMERCELAYER"
	SourceUpdateRequestTypeHubspot         SourceUpdateRequestType = "HUBSPOT"
	SourceUpdateRequestTypeMailgun         SourceUpdateRequestType = "MAILGUN"
	SourceUpdateRequestTypePersona         SourceUpdateRequestType = "PERSONA"
	SourceUpdateRequestTypePipedrive       SourceUpdateRequestType = "PIPEDRIVE"
	SourceUpdateRequestTypeSendgrid        SourceUpdateRequestType = "SENDGRID"
	SourceUpdateRequestTypeWorkos          SourceUpdateRequestType = "WORKOS"
	SourceUpdateRequestTypeSynctera        SourceUpdateRequestType = "SYNCTERA"
	SourceUpdateRequestTypeAwsSns          SourceUpdateRequestType = "AWS_SNS"
	SourceUpdateRequestTypeThreeDEye       SourceUpdateRequestType = "THREE_D_EYE"
	SourceUpdateRequestTypeTwitch          SourceUpdateRequestType = "TWITCH"
	SourceUpdateRequestTypeEnode           SourceUpdateRequestType = "ENODE"
	SourceUpdateRequestTypeFavro           SourceUpdateRequestType = "FAVRO"
	SourceUpdateRequestTypeLinear          SourceUpdateRequestType = "LINEAR"
	SourceUpdateRequestTypeShopline        SourceUpdateRequestType = "SHOPLINE"
	SourceUpdateRequestTypeWix             SourceUpdateRequestType = "WIX"
	SourceUpdateRequestTypeNmi             SourceUpdateRequestType = "NMI"
	SourceUpdateRequestTypeOrb             SourceUpdateRequestType = "ORB"
	SourceUpdateRequestTypePylon           SourceUpdateRequestType = "PYLON"
	SourceUpdateRequestTypeRazorpay        SourceUpdateRequestType = "RAZORPAY"
	SourceUpdateRequestTypeRepay           SourceUpdateRequestType = "REPAY"
	SourceUpdateRequestTypeSquare          SourceUpdateRequestType = "SQUARE"
	SourceUpdateRequestTypeSolidgate       SourceUpdateRequestType = "SOLIDGATE"
	SourceUpdateRequestTypeTrello          SourceUpdateRequestType = "TRELLO"
	SourceUpdateRequestTypeEbay            SourceUpdateRequestType = "EBAY"
	SourceUpdateRequestTypeTelnyx          SourceUpdateRequestType = "TELNYX"
	SourceUpdateRequestTypeDiscord         SourceUpdateRequestType = "DISCORD"
	SourceUpdateRequestTypeTokenio         SourceUpdateRequestType = "TOKENIO"
	SourceUpdateRequestTypeFiserv          SourceUpdateRequestType = "FISERV"
	SourceUpdateRequestTypeBondsmith       SourceUpdateRequestType = "BONDSMITH"
	SourceUpdateRequestTypeVercelLogDrains SourceUpdateRequestType = "VERCEL_LOG_DRAINS"
	SourceUpdateRequestTypeVercel          SourceUpdateRequestType = "VERCEL"
	SourceUpdateRequestTypeTebex           SourceUpdateRequestType = "TEBEX"
	SourceUpdateRequestTypeSlack           SourceUpdateRequestType = "SLACK"
	SourceUpdateRequestTypeMailchimp       SourceUpdateRequestType = "MAILCHIMP"
	SourceUpdateRequestTypePaddle          SourceUpdateRequestType = "PADDLE"
	SourceUpdateRequestTypePaypal          SourceUpdateRequestType = "PAYPAL"
	SourceUpdateRequestTypeTreezor         SourceUpdateRequestType = "TREEZOR"
	SourceUpdateRequestTypePraxis          SourceUpdateRequestType = "PRAXIS"
	SourceUpdateRequestTypeCustomerio      SourceUpdateRequestType = "CUSTOMERIO"
	SourceUpdateRequestTypeFacebook        SourceUpdateRequestType = "FACEBOOK"
	SourceUpdateRequestTypeWhatsapp        SourceUpdateRequestType = "WHATSAPP"
	SourceUpdateRequestTypeReplicate       SourceUpdateRequestType = "REPLICATE"
	SourceUpdateRequestTypeTiktok          SourceUpdateRequestType = "TIKTOK"
	SourceUpdateRequestTypeAirwallex       SourceUpdateRequestType = "AIRWALLEX"
	SourceUpdateRequestTypeZendesk         SourceUpdateRequestType = "ZENDESK"
	SourceUpdateRequestTypeUpollo          SourceUpdateRequestType = "UPOLLO"
	SourceUpdateRequestTypeLinkedin        SourceUpdateRequestType = "LINKEDIN"
)

func NewSourceUpdateRequestTypeFromString(s string) (SourceUpdateRequestType, error) {
	switch s {
	case "WEBHOOK":
		return SourceUpdateRequestTypeWebhook, nil
	case "HTTP":
		return SourceUpdateRequestTypeHttp, nil
	case "MANAGED":
		return SourceUpdateRequestTypeManaged, nil
	case "SANITY":
		return SourceUpdateRequestTypeSanity, nil
	case "BRIDGE":
		return SourceUpdateRequestTypeBridge, nil
	case "CLOUDSIGNAL":
		return SourceUpdateRequestTypeCloudsignal, nil
	case "COURIER":
		return SourceUpdateRequestTypeCourier, nil
	case "FRONTAPP":
		return SourceUpdateRequestTypeFrontapp, nil
	case "ZOOM":
		return SourceUpdateRequestTypeZoom, nil
	case "TWITTER":
		return SourceUpdateRequestTypeTwitter, nil
	case "RECHARGE":
		return SourceUpdateRequestTypeRecharge, nil
	case "STRIPE":
		return SourceUpdateRequestTypeStripe, nil
	case "PROPERTY-FINDER":
		return SourceUpdateRequestTypePropertyFinder, nil
	case "SHOPIFY":
		return SourceUpdateRequestTypeShopify, nil
	case "TWILIO":
		return SourceUpdateRequestTypeTwilio, nil
	case "GITHUB":
		return SourceUpdateRequestTypeGithub, nil
	case "POSTMARK":
		return SourceUpdateRequestTypePostmark, nil
	case "TYPEFORM":
		return SourceUpdateRequestTypeTypeform, nil
	case "XERO":
		return SourceUpdateRequestTypeXero, nil
	case "SVIX":
		return SourceUpdateRequestTypeSvix, nil
	case "ADYEN":
		return SourceUpdateRequestTypeAdyen, nil
	case "AKENEO":
		return SourceUpdateRequestTypeAkeneo, nil
	case "GITLAB":
		return SourceUpdateRequestTypeGitlab, nil
	case "WOOCOMMERCE":
		return SourceUpdateRequestTypeWoocommerce, nil
	case "OURA":
		return SourceUpdateRequestTypeOura, nil
	case "COMMERCELAYER":
		return SourceUpdateRequestTypeCommercelayer, nil
	case "HUBSPOT":
		return SourceUpdateRequestTypeHubspot, nil
	case "MAILGUN":
		return SourceUpdateRequestTypeMailgun, nil
	case "PERSONA":
		return SourceUpdateRequestTypePersona, nil
	case "PIPEDRIVE":
		return SourceUpdateRequestTypePipedrive, nil
	case "SENDGRID":
		return SourceUpdateRequestTypeSendgrid, nil
	case "WORKOS":
		return SourceUpdateRequestTypeWorkos, nil
	case "SYNCTERA":
		return SourceUpdateRequestTypeSynctera, nil
	case "AWS_SNS":
		return SourceUpdateRequestTypeAwsSns, nil
	case "THREE_D_EYE":
		return SourceUpdateRequestTypeThreeDEye, nil
	case "TWITCH":
		return SourceUpdateRequestTypeTwitch, nil
	case "ENODE":
		return SourceUpdateRequestTypeEnode, nil
	case "FAVRO":
		return SourceUpdateRequestTypeFavro, nil
	case "LINEAR":
		return SourceUpdateRequestTypeLinear, nil
	case "SHOPLINE":
		return SourceUpdateRequestTypeShopline, nil
	case "WIX":
		return SourceUpdateRequestTypeWix, nil
	case "NMI":
		return SourceUpdateRequestTypeNmi, nil
	case "ORB":
		return SourceUpdateRequestTypeOrb, nil
	case "PYLON":
		return SourceUpdateRequestTypePylon, nil
	case "RAZORPAY":
		return SourceUpdateRequestTypeRazorpay, nil
	case "REPAY":
		return SourceUpdateRequestTypeRepay, nil
	case "SQUARE":
		return SourceUpdateRequestTypeSquare, nil
	case "SOLIDGATE":
		return SourceUpdateRequestTypeSolidgate, nil
	case "TRELLO":
		return SourceUpdateRequestTypeTrello, nil
	case "EBAY":
		return SourceUpdateRequestTypeEbay, nil
	case "TELNYX":
		return SourceUpdateRequestTypeTelnyx, nil
	case "DISCORD":
		return SourceUpdateRequestTypeDiscord, nil
	case "TOKENIO":
		return SourceUpdateRequestTypeTokenio, nil
	case "FISERV":
		return SourceUpdateRequestTypeFiserv, nil
	case "BONDSMITH":
		return SourceUpdateRequestTypeBondsmith, nil
	case "VERCEL_LOG_DRAINS":
		return SourceUpdateRequestTypeVercelLogDrains, nil
	case "VERCEL":
		return SourceUpdateRequestTypeVercel, nil
	case "TEBEX":
		return SourceUpdateRequestTypeTebex, nil
	case "SLACK":
		return SourceUpdateRequestTypeSlack, nil
	case "MAILCHIMP":
		return SourceUpdateRequestTypeMailchimp, nil
	case "PADDLE":
		return SourceUpdateRequestTypePaddle, nil
	case "PAYPAL":
		return SourceUpdateRequestTypePaypal, nil
	case "TREEZOR":
		return SourceUpdateRequestTypeTreezor, nil
	case "PRAXIS":
		return SourceUpdateRequestTypePraxis, nil
	case "CUSTOMERIO":
		return SourceUpdateRequestTypeCustomerio, nil
	case "FACEBOOK":
		return SourceUpdateRequestTypeFacebook, nil
	case "WHATSAPP":
		return SourceUpdateRequestTypeWhatsapp, nil
	case "REPLICATE":
		return SourceUpdateRequestTypeReplicate, nil
	case "TIKTOK":
		return SourceUpdateRequestTypeTiktok, nil
	case "AIRWALLEX":
		return SourceUpdateRequestTypeAirwallex, nil
	case "ZENDESK":
		return SourceUpdateRequestTypeZendesk, nil
	case "UPOLLO":
		return SourceUpdateRequestTypeUpollo, nil
	case "LINKEDIN":
		return SourceUpdateRequestTypeLinkedin, nil
	}
	var t SourceUpdateRequestType
	return "", fmt.Errorf("%s is not a valid %T", s, t)
}

func (s SourceUpdateRequestType) Ptr() *SourceUpdateRequestType {
	return &s
}

// Type of the source
type SourceUpsertRequestType string

const (
	SourceUpsertRequestTypeWebhook         SourceUpsertRequestType = "WEBHOOK"
	SourceUpsertRequestTypeHttp            SourceUpsertRequestType = "HTTP"
	SourceUpsertRequestTypeManaged         SourceUpsertRequestType = "MANAGED"
	SourceUpsertRequestTypeSanity          SourceUpsertRequestType = "SANITY"
	SourceUpsertRequestTypeBridge          SourceUpsertRequestType = "BRIDGE"
	SourceUpsertRequestTypeCloudsignal     SourceUpsertRequestType = "CLOUDSIGNAL"
	SourceUpsertRequestTypeCourier         SourceUpsertRequestType = "COURIER"
	SourceUpsertRequestTypeFrontapp        SourceUpsertRequestType = "FRONTAPP"
	SourceUpsertRequestTypeZoom            SourceUpsertRequestType = "ZOOM"
	SourceUpsertRequestTypeTwitter         SourceUpsertRequestType = "TWITTER"
	SourceUpsertRequestTypeRecharge        SourceUpsertRequestType = "RECHARGE"
	SourceUpsertRequestTypeStripe          SourceUpsertRequestType = "STRIPE"
	SourceUpsertRequestTypePropertyFinder  SourceUpsertRequestType = "PROPERTY-FINDER"
	SourceUpsertRequestTypeShopify         SourceUpsertRequestType = "SHOPIFY"
	SourceUpsertRequestTypeTwilio          SourceUpsertRequestType = "TWILIO"
	SourceUpsertRequestTypeGithub          SourceUpsertRequestType = "GITHUB"
	SourceUpsertRequestTypePostmark        SourceUpsertRequestType = "POSTMARK"
	SourceUpsertRequestTypeTypeform        SourceUpsertRequestType = "TYPEFORM"
	SourceUpsertRequestTypeXero            SourceUpsertRequestType = "XERO"
	SourceUpsertRequestTypeSvix            SourceUpsertRequestType = "SVIX"
	SourceUpsertRequestTypeAdyen           SourceUpsertRequestType = "ADYEN"
	SourceUpsertRequestTypeAkeneo          SourceUpsertRequestType = "AKENEO"
	SourceUpsertRequestTypeGitlab          SourceUpsertRequestType = "GITLAB"
	SourceUpsertRequestTypeWoocommerce     SourceUpsertRequestType = "WOOCOMMERCE"
	SourceUpsertRequestTypeOura            SourceUpsertRequestType = "OURA"
	SourceUpsertRequestTypeCommercelayer   SourceUpsertRequestType = "COMMERCELAYER"
	SourceUpsertRequestTypeHubspot         SourceUpsertRequestType = "HUBSPOT"
	SourceUpsertRequestTypeMailgun         SourceUpsertRequestType = "MAILGUN"
	SourceUpsertRequestTypePersona         SourceUpsertRequestType = "PERSONA"
	SourceUpsertRequestTypePipedrive       SourceUpsertRequestType = "PIPEDRIVE"
	SourceUpsertRequestTypeSendgrid        SourceUpsertRequestType = "SENDGRID"
	SourceUpsertRequestTypeWorkos          SourceUpsertRequestType = "WORKOS"
	SourceUpsertRequestTypeSynctera        SourceUpsertRequestType = "SYNCTERA"
	SourceUpsertRequestTypeAwsSns          SourceUpsertRequestType = "AWS_SNS"
	SourceUpsertRequestTypeThreeDEye       SourceUpsertRequestType = "THREE_D_EYE"
	SourceUpsertRequestTypeTwitch          SourceUpsertRequestType = "TWITCH"
	SourceUpsertRequestTypeEnode           SourceUpsertRequestType = "ENODE"
	SourceUpsertRequestTypeFavro           SourceUpsertRequestType = "FAVRO"
	SourceUpsertRequestTypeLinear          SourceUpsertRequestType = "LINEAR"
	SourceUpsertRequestTypeShopline        SourceUpsertRequestType = "SHOPLINE"
	SourceUpsertRequestTypeWix             SourceUpsertRequestType = "WIX"
	SourceUpsertRequestTypeNmi             SourceUpsertRequestType = "NMI"
	SourceUpsertRequestTypeOrb             SourceUpsertRequestType = "ORB"
	SourceUpsertRequestTypePylon           SourceUpsertRequestType = "PYLON"
	SourceUpsertRequestTypeRazorpay        SourceUpsertRequestType = "RAZORPAY"
	SourceUpsertRequestTypeRepay           SourceUpsertRequestType = "REPAY"
	SourceUpsertRequestTypeSquare          SourceUpsertRequestType = "SQUARE"
	SourceUpsertRequestTypeSolidgate       SourceUpsertRequestType = "SOLIDGATE"
	SourceUpsertRequestTypeTrello          SourceUpsertRequestType = "TRELLO"
	SourceUpsertRequestTypeEbay            SourceUpsertRequestType = "EBAY"
	SourceUpsertRequestTypeTelnyx          SourceUpsertRequestType = "TELNYX"
	SourceUpsertRequestTypeDiscord         SourceUpsertRequestType = "DISCORD"
	SourceUpsertRequestTypeTokenio         SourceUpsertRequestType = "TOKENIO"
	SourceUpsertRequestTypeFiserv          SourceUpsertRequestType = "FISERV"
	SourceUpsertRequestTypeBondsmith       SourceUpsertRequestType = "BONDSMITH"
	SourceUpsertRequestTypeVercelLogDrains SourceUpsertRequestType = "VERCEL_LOG_DRAINS"
	SourceUpsertRequestTypeVercel          SourceUpsertRequestType = "VERCEL"
	SourceUpsertRequestTypeTebex           SourceUpsertRequestType = "TEBEX"
	SourceUpsertRequestTypeSlack           SourceUpsertRequestType = "SLACK"
	SourceUpsertRequestTypeMailchimp       SourceUpsertRequestType = "MAILCHIMP"
	SourceUpsertRequestTypePaddle          SourceUpsertRequestType = "PADDLE"
	SourceUpsertRequestTypePaypal          SourceUpsertRequestType = "PAYPAL"
	SourceUpsertRequestTypeTreezor         SourceUpsertRequestType = "TREEZOR"
	SourceUpsertRequestTypePraxis          SourceUpsertRequestType = "PRAXIS"
	SourceUpsertRequestTypeCustomerio      SourceUpsertRequestType = "CUSTOMERIO"
	SourceUpsertRequestTypeFacebook        SourceUpsertRequestType = "FACEBOOK"
	SourceUpsertRequestTypeWhatsapp        SourceUpsertRequestType = "WHATSAPP"
	SourceUpsertRequestTypeReplicate       SourceUpsertRequestType = "REPLICATE"
	SourceUpsertRequestTypeTiktok          SourceUpsertRequestType = "TIKTOK"
	SourceUpsertRequestTypeAirwallex       SourceUpsertRequestType = "AIRWALLEX"
	SourceUpsertRequestTypeZendesk         SourceUpsertRequestType = "ZENDESK"
	SourceUpsertRequestTypeUpollo          SourceUpsertRequestType = "UPOLLO"
	SourceUpsertRequestTypeLinkedin        SourceUpsertRequestType = "LINKEDIN"
)

func NewSourceUpsertRequestTypeFromString(s string) (SourceUpsertRequestType, error) {
	switch s {
	case "WEBHOOK":
		return SourceUpsertRequestTypeWebhook, nil
	case "HTTP":
		return SourceUpsertRequestTypeHttp, nil
	case "MANAGED":
		return SourceUpsertRequestTypeManaged, nil
	case "SANITY":
		return SourceUpsertRequestTypeSanity, nil
	case "BRIDGE":
		return SourceUpsertRequestTypeBridge, nil
	case "CLOUDSIGNAL":
		return SourceUpsertRequestTypeCloudsignal, nil
	case "COURIER":
		return SourceUpsertRequestTypeCourier, nil
	case "FRONTAPP":
		return SourceUpsertRequestTypeFrontapp, nil
	case "ZOOM":
		return SourceUpsertRequestTypeZoom, nil
	case "TWITTER":
		return SourceUpsertRequestTypeTwitter, nil
	case "RECHARGE":
		return SourceUpsertRequestTypeRecharge, nil
	case "STRIPE":
		return SourceUpsertRequestTypeStripe, nil
	case "PROPERTY-FINDER":
		return SourceUpsertRequestTypePropertyFinder, nil
	case "SHOPIFY":
		return SourceUpsertRequestTypeShopify, nil
	case "TWILIO":
		return SourceUpsertRequestTypeTwilio, nil
	case "GITHUB":
		return SourceUpsertRequestTypeGithub, nil
	case "POSTMARK":
		return SourceUpsertRequestTypePostmark, nil
	case "TYPEFORM":
		return SourceUpsertRequestTypeTypeform, nil
	case "XERO":
		return SourceUpsertRequestTypeXero, nil
	case "SVIX":
		return SourceUpsertRequestTypeSvix, nil
	case "ADYEN":
		return SourceUpsertRequestTypeAdyen, nil
	case "AKENEO":
		return SourceUpsertRequestTypeAkeneo, nil
	case "GITLAB":
		return SourceUpsertRequestTypeGitlab, nil
	case "WOOCOMMERCE":
		return SourceUpsertRequestTypeWoocommerce, nil
	case "OURA":
		return SourceUpsertRequestTypeOura, nil
	case "COMMERCELAYER":
		return SourceUpsertRequestTypeCommercelayer, nil
	case "HUBSPOT":
		return SourceUpsertRequestTypeHubspot, nil
	case "MAILGUN":
		return SourceUpsertRequestTypeMailgun, nil
	case "PERSONA":
		return SourceUpsertRequestTypePersona, nil
	case "PIPEDRIVE":
		return SourceUpsertRequestTypePipedrive, nil
	case "SENDGRID":
		return SourceUpsertRequestTypeSendgrid, nil
	case "WORKOS":
		return SourceUpsertRequestTypeWorkos, nil
	case "SYNCTERA":
		return SourceUpsertRequestTypeSynctera, nil
	case "AWS_SNS":
		return SourceUpsertRequestTypeAwsSns, nil
	case "THREE_D_EYE":
		return SourceUpsertRequestTypeThreeDEye, nil
	case "TWITCH":
		return SourceUpsertRequestTypeTwitch, nil
	case "ENODE":
		return SourceUpsertRequestTypeEnode, nil
	case "FAVRO":
		return SourceUpsertRequestTypeFavro, nil
	case "LINEAR":
		return SourceUpsertRequestTypeLinear, nil
	case "SHOPLINE":
		return SourceUpsertRequestTypeShopline, nil
	case "WIX":
		return SourceUpsertRequestTypeWix, nil
	case "NMI":
		return SourceUpsertRequestTypeNmi, nil
	case "ORB":
		return SourceUpsertRequestTypeOrb, nil
	case "PYLON":
		return SourceUpsertRequestTypePylon, nil
	case "RAZORPAY":
		return SourceUpsertRequestTypeRazorpay, nil
	case "REPAY":
		return SourceUpsertRequestTypeRepay, nil
	case "SQUARE":
		return SourceUpsertRequestTypeSquare, nil
	case "SOLIDGATE":
		return SourceUpsertRequestTypeSolidgate, nil
	case "TRELLO":
		return SourceUpsertRequestTypeTrello, nil
	case "EBAY":
		return SourceUpsertRequestTypeEbay, nil
	case "TELNYX":
		return SourceUpsertRequestTypeTelnyx, nil
	case "DISCORD":
		return SourceUpsertRequestTypeDiscord, nil
	case "TOKENIO":
		return SourceUpsertRequestTypeTokenio, nil
	case "FISERV":
		return SourceUpsertRequestTypeFiserv, nil
	case "BONDSMITH":
		return SourceUpsertRequestTypeBondsmith, nil
	case "VERCEL_LOG_DRAINS":
		return SourceUpsertRequestTypeVercelLogDrains, nil
	case "VERCEL":
		return SourceUpsertRequestTypeVercel, nil
	case "TEBEX":
		return SourceUpsertRequestTypeTebex, nil
	case "SLACK":
		return SourceUpsertRequestTypeSlack, nil
	case "MAILCHIMP":
		return SourceUpsertRequestTypeMailchimp, nil
	case "PADDLE":
		return SourceUpsertRequestTypePaddle, nil
	case "PAYPAL":
		return SourceUpsertRequestTypePaypal, nil
	case "TREEZOR":
		return SourceUpsertRequestTypeTreezor, nil
	case "PRAXIS":
		return SourceUpsertRequestTypePraxis, nil
	case "CUSTOMERIO":
		return SourceUpsertRequestTypeCustomerio, nil
	case "FACEBOOK":
		return SourceUpsertRequestTypeFacebook, nil
	case "WHATSAPP":
		return SourceUpsertRequestTypeWhatsapp, nil
	case "REPLICATE":
		return SourceUpsertRequestTypeReplicate, nil
	case "TIKTOK":
		return SourceUpsertRequestTypeTiktok, nil
	case "AIRWALLEX":
		return SourceUpsertRequestTypeAirwallex, nil
	case "ZENDESK":
		return SourceUpsertRequestTypeZendesk, nil
	case "UPOLLO":
		return SourceUpsertRequestTypeUpollo, nil
	case "LINKEDIN":
		return SourceUpsertRequestTypeLinkedin, nil
	}
	var t SourceUpsertRequestType
	return "", fmt.Errorf("%s is not a valid %T", s, t)
}

func (s SourceUpsertRequestType) Ptr() *SourceUpsertRequestType {
	return &s
}

type SourceUpdateRequest struct {
	// A unique name for the source
	Name *core.Optional[string] `json:"name,omitempty" url:"-"`
	// Type of the source
	Type *core.Optional[SourceUpdateRequestType] `json:"type,omitempty" url:"-"`
	// Description for the source
	Description *core.Optional[string]           `json:"description,omitempty" url:"-"`
	Config      *core.Optional[SourceTypeConfig] `json:"config,omitempty" url:"-"`
}

type SourceUpsertRequest struct {
	// A unique name for the source
	Name string `json:"name" url:"-"`
	// Type of the source
	Type *core.Optional[SourceUpsertRequestType] `json:"type,omitempty" url:"-"`
	// Description for the source
	Description *core.Optional[string]           `json:"description,omitempty" url:"-"`
	Config      *core.Optional[SourceTypeConfig] `json:"config,omitempty" url:"-"`
}
