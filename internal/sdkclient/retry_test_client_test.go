package sdkclient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// scriptedHTTPClient returns a pre-programmed sequence of responses (one per
// call), and records each request body it observed so tests can confirm the
// body was correctly replayed across retries.
type scriptedHTTPClient struct {
	mu        sync.Mutex
	responses []*http.Response
	errors    []error
	bodies    []string
	calls     int
}

func (s *scriptedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.calls
	s.calls++
	body := ""
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		body = string(b)
	}
	s.bodies = append(s.bodies, body)
	if idx < len(s.errors) && s.errors[idx] != nil {
		return nil, s.errors[idx]
	}
	if idx >= len(s.responses) {
		return nil, errors.New("scriptedHTTPClient: no more responses programmed")
	}
	return s.responses[idx], nil
}

func resp(status int, retryAfter string) *http.Response {
	h := http.Header{}
	if retryAfter != "" {
		h.Set("Retry-After", retryAfter)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
}

func newReq(t *testing.T, body string) *http.Request {
	t.Helper()
	r, err := http.NewRequest("POST", "https://api.example/test", strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return r
}

func newRetryClientForTest(under HTTPDoer, maxRetries int, sleeps *[]time.Duration) *retrying429Client {
	c := newRetrying429Client(under, maxRetries)
	c.sleep = func(d time.Duration) { *sleeps = append(*sleeps, d) }
	return c
}

func TestRetrying429Client_SuccessFirstAttempt(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{resp(200, "")},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	r, err := c.Do(newReq(t, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", r.StatusCode)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}
	if len(sleeps) != 0 {
		t.Fatalf("expected no sleeps, got %v", sleeps)
	}
}

func TestRetrying429Client_RetriesUntilSuccess(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{
			resp(429, "1"),
			resp(429, "2"),
			resp(200, ""),
		},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	r, err := c.Do(newReq(t, `{"a":1}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", r.StatusCode)
	}
	if mock.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.calls)
	}
	if want := []time.Duration{1 * time.Second, 2 * time.Second}; len(sleeps) != 2 || sleeps[0] != want[0] || sleeps[1] != want[1] {
		t.Fatalf("expected sleeps %v, got %v", want, sleeps)
	}
	// Body should have been replayed identically on every call.
	for i, b := range mock.bodies {
		if b != `{"a":1}` {
			t.Fatalf("call %d: expected body to be replayed, got %q", i, b)
		}
	}
}

func TestRetrying429Client_ReturnsLast429AfterMaxRetries(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{
			resp(429, "1"),
			resp(429, "1"),
			resp(429, "1"),
		},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 2, &sleeps)

	r, err := c.Do(newReq(t, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != 429 {
		t.Fatalf("expected final 429, got %d", r.StatusCode)
	}
	// maxRetries=2 means we make up to 3 attempts total, sleeping twice between.
	if mock.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("expected 2 sleeps, got %v", sleeps)
	}
}

func TestRetrying429Client_ExponentialBackoffWhenNoRetryAfter(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{
			resp(429, ""),
			resp(429, ""),
			resp(429, ""),
			resp(200, ""),
		},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	if _, err := c.Do(newReq(t, "")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2s, 4s, 8s
	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}
	if len(sleeps) != len(want) {
		t.Fatalf("expected %d sleeps, got %v", len(want), sleeps)
	}
	for i, w := range want {
		if sleeps[i] != w {
			t.Fatalf("sleep[%d] expected %v, got %v", i, w, sleeps[i])
		}
	}
}

func TestRetrying429Client_PropagatesNon429(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{resp(500, "")},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	r, err := c.Do(newReq(t, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.StatusCode != 500 {
		t.Fatalf("expected 500 propagated, got %d", r.StatusCode)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call (no retry on non-429), got %d", mock.calls)
	}
}

func TestRetrying429Client_PropagatesTransportError(t *testing.T) {
	mock := &scriptedHTTPClient{
		errors:    []error{errors.New("network down")},
		responses: []*http.Response{nil},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	if _, err := c.Do(newReq(t, "")); err == nil || err.Error() != "network down" {
		t.Fatalf("expected transport error, got %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call (no retry on transport error), got %d", mock.calls)
	}
}

func TestRetrying429Client_CapsWaitAt60s(t *testing.T) {
	mock := &scriptedHTTPClient{
		responses: []*http.Response{
			resp(429, "300"), // 5 minute Retry-After
			resp(200, ""),
		},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)

	if _, err := c.Do(newReq(t, "")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sleeps) != 1 || sleeps[0] != 60*time.Second {
		t.Fatalf("expected single 60s sleep, got %v", sleeps)
	}
}

func TestRetrying429Client_ParseRetryAfterHTTPDate(t *testing.T) {
	fixedNow := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	retryAt := fixedNow.Add(7 * time.Second).UTC().Format(http.TimeFormat)
	mock := &scriptedHTTPClient{
		responses: []*http.Response{
			resp(429, retryAt),
			resp(200, ""),
		},
	}
	var sleeps []time.Duration
	c := newRetryClientForTest(mock, 3, &sleeps)
	c.now = func() time.Time { return fixedNow }

	if _, err := c.Do(newReq(t, "")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Allow 1s tolerance for the HTTP date format's second-level precision.
	if len(sleeps) != 1 || sleeps[0] < 6*time.Second || sleeps[0] > 8*time.Second {
		t.Fatalf("expected ~7s sleep, got %v", sleeps)
	}
}
