package sdkclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Client is the main client struct exposed to the provider.
type Client struct {
	RawClient RawClientInterface
}

const (
	defaultAPIBase = "api.hookdeck.com"
)

// RawClientInterface defines the contract for sending requests.
type RawClientInterface interface {
	SendRequest(ctx context.Context, method, path string, opts *RequestOptions) (*http.Response, error)
}

// RateLimiter interface allows for custom rate limiting implementations.
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// RawClient is the concrete implementation with rate limiting.
type RawClient struct {
	apiKey      string
	apiBase     string
	header      http.Header
	httpClient  HTTPDoer
	rateLimiter RateLimiter
}

// HTTPDoer allows for mocking HTTP client in tests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RequestOptions contains optional parameters for requests.
type RequestOptions struct {
	Body        io.Reader
	Headers     http.Header
	QueryParams url.Values
}

// RawClientOption allows configuring the RawClient.
type RawClientOption func(*RawClient)

// WithHTTPClient sets a custom HTTP client (useful for testing).
func WithHTTPClient(client HTTPDoer) RawClientOption {
	return func(c *RawClient) {
		c.httpClient = client
	}
}

// WithRateLimiter sets a custom rate limiter.
func WithRateLimiter(limiter RateLimiter) RawClientOption {
	return func(c *RawClient) {
		c.rateLimiter = limiter
	}
}

// HookdeckRateLimiter implements the Hookdeck API rate limit (240 requests per minute).
type HookdeckRateLimiter struct {
	limiter *rate.Limiter
}

// NewHookdeckRateLimiter creates a rate limiter for Hookdeck's API limits.
func NewHookdeckRateLimiter() *HookdeckRateLimiter {
	// 240 requests per minute = 4 per second
	// Use rate.Every for cleaner per-minute expression
	// Burst of 10 allows for short bursts of activity
	return &HookdeckRateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Minute/240), 10),
	}
}

// Wait implements the RateLimiter interface.
func (h *HookdeckRateLimiter) Wait(ctx context.Context) error {
	return h.limiter.Wait(ctx)
}

// InitHookdeckSDKClient creates a client with Hookdeck's rate limiting.
func InitHookdeckSDKClient(apiBase string, apiKey string, providerVersion string, opts ...RawClientOption) Client {
	if apiBase == "" {
		apiBase = defaultAPIBase
	}
	if !strings.HasPrefix(apiBase, "http") {
		apiBase = "https://" + apiBase
	}
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)

	// Create RawClient with Hookdeck rate limiting by default
	rawClient := &RawClient{
		apiKey:      apiKey,
		apiBase:     apiBase,
		header:      header,
		httpClient:  http.DefaultClient,
		rateLimiter: NewHookdeckRateLimiter(),
	}

	// Apply any options
	for _, opt := range opts {
		opt(rawClient)
	}

	return Client{RawClient: rawClient}
}

// SendRequest sends an HTTP request with rate limiting.
func (c *RawClient) SendRequest(ctx context.Context, method, path string, opts *RequestOptions) (*http.Response, error) {
	// IMPORTANT: Context Handling Issue with Terraform
	//
	// We accept a context parameter to maintain interface compatibility, but we DO NOT use it
	// for the HTTP request. This is because Terraform's acceptance test framework and some
	// provider operations (especially post-apply refresh) cancel contexts too aggressively,
	// which causes requests to fail with "context canceled" errors.
	//
	// By not passing the context to http.NewRequest, we ensure that:
	// 1. API requests can complete even when Terraform cancels the context
	// 2. Rate limiting is still properly enforced
	// 3. The provider remains stable during Terraform operations
	//
	// This is a deliberate design decision after debugging context cancellation issues
	// in acceptance tests. The trade-off is that we lose the ability to cancel in-flight
	// requests, but this is acceptable for Terraform provider operations which need to
	// complete for state consistency.
	//
	// TODO: Future Improvements
	// - Investigate the root cause of aggressive context cancellation in Terraform
	// - Consider implementing a grace period mechanism for context cancellation
	// - Explore whether newer versions of Terraform handle contexts differently
	// - Potentially add configurable behavior for context handling
	_ = ctx // Explicitly ignore the context parameter

	// Apply rate limiting if configured
	if c.rateLimiter != nil {
		// Use context.Background() for rate limiting to ensure it's never canceled
		// and we always respect API rate limits
		if err := c.rateLimiter.Wait(context.Background()); err != nil {
			return nil, fmt.Errorf("rate limit wait failed: %w", err)
		}
	}

	url := fmt.Sprintf("%s%s", c.apiBase, path)
	if opts != nil && opts.QueryParams != nil {
		url += "?" + opts.QueryParams.Encode()
	}

	var body io.Reader
	if opts != nil {
		body = opts.Body
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set default headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	for key, value := range c.header {
		for _, v := range value {
			req.Header.Add(key, v)
		}
	}

	// Set custom headers if provided
	if opts != nil && opts.Headers != nil {
		for key, values := range opts.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	return c.httpClient.Do(req)
}
