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
//
// Rate Limiting Strategy:
// - Initial burst: 5 requests (allows immediate execution for small operations)
// - Steady rate: 230 requests/minute (3.83 req/sec)
// - This results in max 235 requests in the first minute, then 230/min thereafter
// - Conservative approach ensures we stay under the 240 req/min API limit
//
// API Rate Limit Headers:
// The Hookdeck API returns the following rate limit headers in responses:
// - Retry-After: Seconds before next request can be made
// - X-RateLimit-Limit: Requests per minute limit for the API key
// - X-RateLimit-Remaining: Remaining requests for the rate limit period
// - X-RateLimit-Reset: ISO timestamp of the next rate limit period reset
// Currently, we use client-side rate limiting and don't dynamically adjust based
// on these headers, but they're available for debugging and future enhancements.
//
// Retry Considerations:
// - We intentionally DO NOT implement automatic retries on 429 errors
// - Retries would consume additional tokens from our rate limit bucket,
//   potentially causing cascading failures for subsequent requests
// - If retry logic is needed, implement using recursion with retry count tracking,
//   ensuring each retry attempt goes through the rate limiter to maintain
//   the global rate limit invariant
//
// Performance Trade-offs:
// - This rate limiter works well for large Terraform states (200+ resources)
// - However, it significantly impacts small deployments:
//   - 100 resources now take ~30 seconds vs ~5 seconds previously
//   - This is an intentional trade-off for reliability and compliance
// - Future optimization could implement a true "240 requests per sliding minute"
//   window instead of the current token bucket approach
// - Potential solution: In-memory FIFO queue that tracks the last minute's
//   requests and blocks when 240 requests have been made in the past 60 seconds
func NewHookdeckRateLimiter() *HookdeckRateLimiter {
	// API limit: 240 requests per minute
	// We use 230 req/min with burst of 5 to stay safely under the limit
	// 230 + 5 = 235 total possible requests in first minute (5 request buffer)
	return &HookdeckRateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Minute/230), 5),
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
	// IMPORTANT: Context Handling in Acceptance Tests
	//
	// We accept a context parameter to maintain interface compatibility, but we DO NOT use it
	// for the HTTP request. This is specifically to work around an issue with Terraform's
	// acceptance test framework, which cancels contexts aggressively during certain test
	// operations (particularly during post-apply refresh), causing requests to fail with
	// "context canceled" errors.
	//
	// Trade-off: We lose the ability to cancel in-flight requests, but this is acceptable
	// because:
	// - We're only making simple API calls without complex cancellation logic
	// - The production provider currently doesn't pass context to HTTP requests either
	// - Terraform operations need to complete for state consistency
	//
	// This matches the existing production behavior, so we're not losing any functionality.
	//
	// TODO: Understand acceptance test framework context cancellation and make it work properly
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
