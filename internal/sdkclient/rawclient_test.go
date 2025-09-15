package sdkclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"fmt"
	"strings"
)

// MockRateLimiter for testing rate limiting behavior.
type MockRateLimiter struct {
	waitCalls int32
	waitDelay time.Duration
	waitError error
	mu        sync.Mutex // Add mutex for sequential enforcement
}

func (m *MockRateLimiter) Wait(ctx context.Context) error {
	// Lock to ensure sequential processing when testing rate limiting
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt32(&m.waitCalls, 1)
	if m.waitDelay > 0 {
		select {
		case <-time.After(m.waitDelay):
			// Delay completed
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.waitError
}

func (m *MockRateLimiter) GetWaitCalls() int {
	return int(atomic.LoadInt32(&m.waitCalls))
}

// TrackingMockHTTPClient tracks when requests are made.
type TrackingMockHTTPClient struct {
	mu            sync.Mutex
	requestTimes  []time.Time
	requests      []*http.Request
	requestDelay  time.Duration
	responseCode  int
	responseError error
}

func NewTrackingMockHTTPClient() *TrackingMockHTTPClient {
	return &TrackingMockHTTPClient{
		requestTimes: make([]time.Time, 0),
		requests:     make([]*http.Request, 0),
		responseCode: 200,
	}
}

func (m *TrackingMockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	m.requestTimes = append(m.requestTimes, time.Now())
	m.requests = append(m.requests, req)
	delay := m.requestDelay
	m.mu.Unlock()

	// Simulate request processing time
	if delay > 0 {
		time.Sleep(delay)
	}

	if m.responseError != nil {
		return nil, m.responseError
	}

	return &http.Response{
		StatusCode: m.responseCode,
		Body:       io.NopCloser(bytes.NewBufferString(`{"success": true}`)),
		Header:     make(http.Header),
	}, nil
}

func (m *TrackingMockHTTPClient) GetRequestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

func (m *TrackingMockHTTPClient) GetRequestTimes() []time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	times := make([]time.Time, len(m.requestTimes))
	copy(times, m.requestTimes)
	return times
}

func TestRawClient_WithRateLimiter(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	mockRateLimiter := &MockRateLimiter{
		waitDelay: 100 * time.Millisecond, // Each Wait() takes 100ms
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	// Send 5 requests
	for i := 0; i < 5; i++ {
		_, err := rawClient.SendRequest(context.Background(), "GET", "/test", nil)
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	// Verify rate limiter was called for each request
	if calls := mockRateLimiter.GetWaitCalls(); calls != 5 {
		t.Errorf("Expected 5 rate limiter calls, got %d", calls)
	}

	// Verify all requests were made
	if count := mockHTTP.GetRequestCount(); count != 5 {
		t.Errorf("Expected 5 requests, got %d", count)
	}
}

func TestRawClient_NoRateLimiter(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	mockHTTP.requestDelay = 50 * time.Millisecond // Each request takes 50ms

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: nil, // No rate limiting
	}

	numRequests := 10
	var wg sync.WaitGroup
	startTime := time.Now()

	// Send all requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rawClient.SendRequest(context.Background(), "GET", "/test", nil)
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Without rate limiting, all requests should run concurrently
	// Total time should be ~50ms (parallel) not 500ms (sequential)
	if totalTime > 200*time.Millisecond {
		t.Errorf("Concurrent requests took too long: %v (expected ~50ms for parallel execution)", totalTime)
	}

	// Verify all requests were made
	if got := mockHTTP.GetRequestCount(); got != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, got)
	}
}

func TestRawClient_RateLimiterBlocking(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	mockRateLimiter := &MockRateLimiter{
		waitDelay: 200 * time.Millisecond, // Slow but reasonable rate limiter
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	// Test that rate limiter blocks request until ready
	start := time.Now()

	// Make a synchronous request
	_, err := rawClient.SendRequest(context.Background(), "GET", "/test", nil)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}

	elapsed := time.Since(start)

	// Verify the rate limiter was called
	if calls := mockRateLimiter.GetWaitCalls(); calls != 1 {
		t.Errorf("Expected 1 rate limiter call, got %d", calls)
	}

	// Verify request was made after rate limiter delay
	if count := mockHTTP.GetRequestCount(); count != 1 {
		t.Errorf("Expected 1 request, got %d", count)
	}

	// Verify it took at least the rate limiter delay
	if elapsed < 200*time.Millisecond {
		t.Errorf("Request completed too quickly: %v (rate limiter not working)", elapsed)
	}

	// Verify it didn't take too long
	if elapsed > 300*time.Millisecond {
		t.Errorf("Request took too long: %v", elapsed)
	}
}

func TestRawClient_RequestConstruction(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()

	rawClient := &RawClient{
		apiKey:      "test-api-key",
		apiBase:     "https://api.hookdeck.com",
		header:      http.Header{"X-Provider-Version": []string{"1.0.0"}},
		httpClient:  mockHTTP,
		rateLimiter: nil, // No rate limiting for this test
	}

	opts := &RequestOptions{
		Body: bytes.NewBufferString(`{"test": "data"}`),
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
		QueryParams: map[string][]string{
			"foo": {"bar"},
			"baz": {"qux"},
		},
	}

	_, err := rawClient.SendRequest(context.Background(), "POST", "/test", opts)
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}

	// Check the constructed request
	reqs := mockHTTP.requests
	if len(reqs) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(reqs))
	}

	req := reqs[0]

	// Check method
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}

	// Check URL and query params
	if req.URL.Query().Get("foo") != "bar" {
		t.Errorf("Missing query param foo=bar")
	}
	if req.URL.Query().Get("baz") != "qux" {
		t.Errorf("Missing query param baz=qux")
	}

	// Check headers
	if req.Header.Get("Authorization") != "Bearer test-api-key" {
		t.Errorf("Wrong Authorization header: %s", req.Header.Get("Authorization"))
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Wrong Content-Type: %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("X-Provider-Version") != "1.0.0" {
		t.Errorf("Missing provider version header")
	}
}

func TestRawClient_ConcurrentRequestsWithRateLimiter(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	mockRateLimiter := &MockRateLimiter{
		waitDelay: 50 * time.Millisecond, // Each wait takes 50ms
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	numRequests := 10
	var wg sync.WaitGroup
	var successCount int32

	startTime := time.Now()

	// Send requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := rawClient.SendRequest(context.Background(), "GET", "/test", nil)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// With rate limiter delay, requests should be serialized
	// 10 requests * 50ms = 500ms minimum
	expectedMinTime := 450 * time.Millisecond // Allow some tolerance
	if totalTime < expectedMinTime {
		t.Errorf("Requests completed too quickly: %v < %v (rate limiting not working)", totalTime, expectedMinTime)
	}

	// Verify all requests succeeded
	if atomic.LoadInt32(&successCount) != int32(numRequests) {
		t.Errorf("Not all requests succeeded: %d/%d", successCount, numRequests)
	}

	// Verify rate limiter was called for each request
	if calls := mockRateLimiter.GetWaitCalls(); calls != numRequests {
		t.Errorf("Expected %d rate limiter calls, got %d", numRequests, calls)
	}
}

func TestRawClient_RateLimiterEnforcesSequentialExecution(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	mockRateLimiter := &MockRateLimiter{
		waitDelay: 100 * time.Millisecond, // Each wait takes 100ms
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	numRequests := 5
	var wg sync.WaitGroup

	// Send requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rawClient.SendRequest(context.Background(), "GET", "/test", nil)
		}()
	}

	wg.Wait()

	// Get actual request times from the HTTP client (after rate limiting)
	requestTimes := mockHTTP.GetRequestTimes()

	// Check that requests were properly spaced due to rate limiting
	for i := 1; i < len(requestTimes); i++ {
		gap := requestTimes[i].Sub(requestTimes[i-1])
		// With 100ms rate limiter delay, gaps should be at least 90ms (allowing some tolerance)
		if gap < 90*time.Millisecond {
			t.Errorf("Requests %d and %d were made too close together: %v (rate limiting not enforced)", i-1, i, gap)
		}
	}

	// Verify all requests were made
	if len(requestTimes) != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, len(requestTimes))
	}
}

func TestRawClient_RateLimiterErrorHandling(t *testing.T) {
	mockHTTP := NewTrackingMockHTTPClient()
	testError := fmt.Errorf("rate limit error")
	mockRateLimiter := &MockRateLimiter{
		waitError: testError,
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	_, err := rawClient.SendRequest(context.Background(), "GET", "/test", nil)
	if err == nil {
		t.Error("Expected error from rate limiter, got nil")
	}

	// Verify error message contains rate limit context
	if !strings.Contains(err.Error(), "rate limit wait failed") {
		t.Errorf("Error should mention rate limit failure, got: %v", err)
	}

	// Verify request was not made when rate limiter fails
	if mockHTTP.GetRequestCount() != 0 {
		t.Error("Request should not be made when rate limiter fails")
	}
}

func TestRawClient_RequestThroughputWithRateLimiter(t *testing.T) {
	// Test that rate limiter properly controls request throughput
	mockHTTP := NewTrackingMockHTTPClient()
	requestDelay := 20 * time.Millisecond
	mockRateLimiter := &MockRateLimiter{
		waitDelay: requestDelay,
	}

	rawClient := &RawClient{
		apiKey:      "test-key",
		apiBase:     "https://api.test.com",
		header:      make(http.Header),
		httpClient:  mockHTTP,
		rateLimiter: mockRateLimiter,
	}

	numRequests := 20
	var wg sync.WaitGroup

	startTime := time.Now()

	// Send requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = rawClient.SendRequest(context.Background(), "GET", "/test", nil)
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// Calculate expected time: numRequests * requestDelay
	expectedTime := time.Duration(numRequests) * requestDelay
	tolerance := 100 * time.Millisecond

	if elapsed < expectedTime-tolerance {
		t.Errorf("Requests completed too quickly: %v (expected ~%v)", elapsed, expectedTime)
	}

	// Verify the throughput rate
	throughput := float64(numRequests) / elapsed.Seconds()
	expectedThroughput := 1.0 / requestDelay.Seconds()

	// Allow 20% tolerance for throughput
	if throughput > expectedThroughput*1.2 {
		t.Errorf("Throughput too high: %.2f req/s (expected ~%.2f req/s)", throughput, expectedThroughput)
	}
}
