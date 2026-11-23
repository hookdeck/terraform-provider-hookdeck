package sdkclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"terraform-provider-hookdeck/internal/projectscope"
)

// Client is the main client struct exposed to the provider.
type Client struct {
	RawClient RawClientInterface

	OutpostRawClient RawClientInterface

	// Scope is the key kind and the provider's project.
	Scope projectscope.Scope

	projectChecks *projectChecks
}

const (
	defaultAPIBase        = "api.hookdeck.com"
	defaultOutpostAPIBase = "api.outpost.hookdeck.com"

	// APIVersion is the Hookdeck API version every request targets.
	APIVersion = "2026-09-01"
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
	// writeTimeout bounds a write request; zero means defaultWriteTimeout.
	writeTimeout time.Duration
}

// defaultWriteTimeout bounds one write (POST, PUT, PATCH, DELETE). Writes
// ignore cancellation of the operation's context (see SendRequest), so this
// limit is what ends a write the server never answers. It is far above what a
// healthy write takes, project creation included, so it only ever ends a
// hung connection.
const defaultWriteTimeout = 5 * time.Minute

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
//   - The client does not retry a 429 by itself: a retry takes another token
//     from the rate limit bucket, which can make later requests fail in turn
//   - The Outpost config calls are the exception. The wait for a new project
//     and the config write retry 429 and 5xx a bounded number of times, after
//     a fixed delay (Retry-After is not read)
//   - Every retry goes through the rate limiter, so the global rate limit
//     still holds
//
// Performance Trade-offs:
// - This rate limiter works well for large Terraform states (200+ resources)
// - However, it significantly impacts small deployments:
//   - 100 resources now take ~30 seconds vs ~5 seconds previously
//   - This is an intentional trade-off for reliability and compliance
//   - Future optimization could implement a true "240 requests per sliding minute"
//     window instead of the current token bucket approach
//   - Potential solution: In-memory FIFO queue that tracks the last minute's
//     requests and blocks when 240 requests have been made in the past 60 seconds
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

// Acceptance tests run in parallel, so the test limiter allows more than the
// production one (override with HOOKDECK_TEST_REQUESTS_PER_MINUTE).
const (
	testRequestsPerMinute = 2000
	testRateBurst         = 20
)

// sharedTestRateLimiter, when running under TF_ACC=1, is shared across every
// RawClient created in the process. Each acceptance test step spins up a
// fresh provider (and would otherwise get its own burst); stacking those
// bursts blows past the API's per-key limit over its fixed one-minute window.
// Sharing keeps the cumulative request rate within bounds.
var (
	sharedTestRateLimiterOnce sync.Once
	sharedTestRateLimiterVal  RateLimiter
)

func getSharedTestRateLimiter() RateLimiter {
	sharedTestRateLimiterOnce.Do(func() {
		perMinute := testRequestsPerMinute
		if v, err := strconv.Atoi(os.Getenv("HOOKDECK_TEST_REQUESTS_PER_MINUTE")); err == nil && v > 0 {
			perMinute = v
		}
		sharedTestRateLimiterVal = &HookdeckRateLimiter{
			limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(perMinute)), testRateBurst),
		}
	})
	return sharedTestRateLimiterVal
}

// InitHookdeckSDKClient creates a client with Hookdeck's rate limiting.
func InitHookdeckSDKClient(apiBase, outpostAPIBase, apiKey, providerVersion string, opts ...RawClientOption) Client {
	apiBase = baseURL(apiBase, defaultAPIBase)
	outpostAPIBase = baseURL(outpostAPIBase, defaultOutpostAPIBase)
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)

	limiter := RateLimiter(NewHookdeckRateLimiter())
	httpClient := HTTPDoer(http.DefaultClient)
	if os.Getenv("TF_ACC") == "1" {
		limiter = getSharedTestRateLimiter()
		// Tests spin up many short-lived providers against a single API key;
		// cumulative bursts can still occasionally trip the API's sliding-
		// window 429 even with the shared limiter. Wrap the HTTP client to
		// retry on 429 with the Retry-After header respected. Production
		// behavior intentionally does NOT retry (see comment on
		// HookdeckRateLimiter); this only applies under TF_ACC=1.
		httpClient = newRetrying429Client(http.DefaultClient, 5)
	}

	// Create RawClient with Hookdeck rate limiting by default
	rawClient := &RawClient{
		apiKey:      apiKey,
		apiBase:     apiBase,
		header:      header,
		httpClient:  httpClient,
		rateLimiter: limiter,
	}

	// Apply any options
	for _, opt := range opts {
		opt(rawClient)
	}

	outpostRawClient := &RawClient{
		apiKey:      apiKey,
		apiBase:     outpostAPIBase,
		header:      header,
		httpClient:  rawClient.httpClient,
		rateLimiter: rawClient.rateLimiter,
	}

	return Client{
		RawClient:        rawClient,
		OutpostRawClient: outpostRawClient,
		Scope:            projectscope.Scope{KeyKind: projectscope.KindOfKey(apiKey)},
		projectChecks:    &projectChecks{results: map[string]*projectCheck{}},
	}
}

func baseURL(base, fallback string) string {
	if base == "" {
		base = fallback
	}
	if !strings.HasPrefix(base, "http") {
		base = "https://" + base
	}
	return base
}

// SendRequest sends an HTTP request with rate limiting.
//
// A deadline or cancellation of ctx stops the rate limiter wait and a GET in
// flight. A write runs on a context detached from ctx once it is sent: when
// Terraform is interrupted it cancels the operation's context while the API
// may already have applied the write, and abandoning the request would leave
// the object in Hookdeck but not in state. The write is bounded by
// writeTimeout instead.
func (c *RawClient) SendRequest(ctx context.Context, method, path string, opts *RequestOptions) (*http.Response, error) {
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			// The limiter refuses a wait that would pass the deadline before
			// the deadline arrives; report it as the deadline.
			if _, hasDeadline := ctx.Deadline(); hasDeadline && ctx.Err() == nil {
				err = fmt.Errorf("%w (%v)", context.DeadlineExceeded, err)
			}
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

	reqCtx, cancel := ctx, context.CancelFunc(func() {})
	if method != http.MethodGet && method != http.MethodHead {
		timeout := c.writeTimeout
		if timeout <= 0 {
			timeout = defaultWriteTimeout
		}
		reqCtx, cancel = context.WithTimeout(context.WithoutCancel(ctx), timeout)
	}

	req, err := http.NewRequestWithContext(reqCtx, method, url, body)
	if err != nil {
		cancel()
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}
	// The response body is read after this returns, so the write's time limit
	// stays in force until the body is closed or the limit passes.
	if resp.Body == nil {
		resp.Body = http.NoBody
	}
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

// cancelOnClose releases a request context when the response body is closed.
type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelOnClose) Close() error {
	err := b.ReadCloser.Close()
	b.cancel()
	return err
}
