// This file was auto-generated by Fern from our API Definition.

package transformation

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

func (c *Client) List(
	ctx context.Context,
	request *hookdeckgosdk.TransformationListRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.TransformationPaginatedResult, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := baseURL + "/transformations"
	queryParams, err := internal.QueryValues(request)
	if err != nil {
		return nil, err
	}
	if len(queryParams) > 0 {
		endpointURL += "?" + queryParams.Encode()
	}
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.TransformationPaginatedResult
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Create(
	ctx context.Context,
	request *hookdeckgosdk.TransformationCreateRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.Transformation, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := baseURL + "/transformations"
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	headers.Set("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.Transformation
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodPost,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Request:         request,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Upsert(
	ctx context.Context,
	request *hookdeckgosdk.TransformationUpsertRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.Transformation, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := baseURL + "/transformations"
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	headers.Set("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.Transformation
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
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Retrieve(
	ctx context.Context,
	id string,
	opts ...option.RequestOption,
) (*hookdeckgosdk.Transformation, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/transformations/%v",
		id,
	)
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &hookdeckgosdk.NotFoundError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.Transformation
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Update(
	ctx context.Context,
	id string,
	request *hookdeckgosdk.TransformationUpdateRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.Transformation, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/transformations/%v",
		id,
	)
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	headers.Set("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		404: func(apiError *core.APIError) error {
			return &hookdeckgosdk.NotFoundError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.Transformation
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
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Delete(
	ctx context.Context,
	id string,
	opts ...option.RequestOption,
) (*hookdeckgosdk.TransformationDeleteResponse, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/transformations/%v",
		id,
	)
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &hookdeckgosdk.NotFoundError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.TransformationDeleteResponse
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodDelete,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) Run(
	ctx context.Context,
	request *hookdeckgosdk.TransformationRunRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.TransformationExecutorOutput, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := baseURL + "/transformations/run"
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	headers.Set("Content-Type", "application/json")
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.TransformationExecutorOutput
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
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) ListExecution(
	ctx context.Context,
	id string,
	request *hookdeckgosdk.TransformationListExecutionRequest,
	opts ...option.RequestOption,
) (*hookdeckgosdk.TransformationExecutionPaginatedResult, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/transformations/%v/executions",
		id,
	)
	queryParams, err := internal.QueryValues(request)
	if err != nil {
		return nil, err
	}
	if len(queryParams) > 0 {
		endpointURL += "?" + queryParams.Encode()
	}
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		400: func(apiError *core.APIError) error {
			return &hookdeckgosdk.BadRequestError{
				APIError: apiError,
			}
		},
		422: func(apiError *core.APIError) error {
			return &hookdeckgosdk.UnprocessableEntityError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.TransformationExecutionPaginatedResult
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}

func (c *Client) RetrieveExecution(
	ctx context.Context,
	id string,
	executionId string,
	opts ...option.RequestOption,
) (*hookdeckgosdk.TransformationExecution, error) {
	options := core.NewRequestOptions(opts...)
	baseURL := internal.ResolveBaseURL(
		options.BaseURL,
		c.baseURL,
		"https://api.hookdeck.com/2025-01-01",
	)
	endpointURL := internal.EncodeURL(
		baseURL+"/transformations/%v/executions/%v",
		id,
		executionId,
	)
	headers := internal.MergeHeaders(
		c.header.Clone(),
		options.ToHeader(),
	)
	errorCodes := internal.ErrorCodes{
		404: func(apiError *core.APIError) error {
			return &hookdeckgosdk.NotFoundError{
				APIError: apiError,
			}
		},
	}

	var response *hookdeckgosdk.TransformationExecution
	if err := c.caller.Call(
		ctx,
		&internal.CallParams{
			URL:             endpointURL,
			Method:          http.MethodGet,
			Headers:         headers,
			MaxAttempts:     options.MaxAttempts,
			BodyProperties:  options.BodyProperties,
			QueryParameters: options.QueryParameters,
			Client:          options.HTTPClient,
			Response:        &response,
			ErrorDecoder:    internal.NewErrorDecoder(errorCodes),
		},
	); err != nil {
		return nil, err
	}
	return response, nil
}
