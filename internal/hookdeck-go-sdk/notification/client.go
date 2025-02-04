// This file was auto-generated by Fern from our API Definition.

package notification

import (
	context "context"
	hookdeckgosdk "github.com/hookdeck/hookdeck-go-sdk"
	core "github.com/hookdeck/hookdeck-go-sdk/core"
	internal "github.com/hookdeck/hookdeck-go-sdk/internal"
	option "github.com/hookdeck/hookdeck-go-sdk/option"
	http "net/http"
)

type Client struct {
	baseURL string
	caller  *internal.Caller
	header  http.Header
}

func NewClient(opts ...option.RequestOption) *Client {
	options := core.NewRequestOptions(opts...)
	return &Client{
		baseURL: options.BaseURL,
		caller: internal.NewCaller(
			&internal.CallerParams{
				Client:      options.HTTPClient,
				MaxAttempts: options.MaxAttempts,
			},
		),
		header: options.ToHeader(),
	}
}

func (c *Client) Update(
	ctx context.Context,
	request *hookdeckgosdk.NotificationUpdateRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.ToggleWebhookNotifications, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := baseURL + "/notifications/webhooks"
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	headers.Set("Content-Type", "application/json")

	var response *hookdeckgosdk.ToggleWebhookNotifications
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPut,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}
