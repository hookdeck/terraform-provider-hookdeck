package sdkclient

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"
)

// retrying429Client wraps an HTTPDoer and retries on 429 Too Many Requests,
// respecting the Retry-After header when present. Used in acceptance tests
// only — production behavior intentionally does not retry on 429 (see
// sdkclient.go for that rationale). The retry wrapper exists to absorb
// transient bursts caused by the test framework spinning up many short-lived
// providers in sequence against a single API key.
type retrying429Client struct {
	underlying HTTPDoer
	maxRetries int
	// sleep is overridable for tests. Defaults to time.Sleep.
	sleep func(time.Duration)
	// now is overridable for tests. Defaults to time.Now.
	now func() time.Time
}

func newRetrying429Client(underlying HTTPDoer, maxRetries int) *retrying429Client {
	return &retrying429Client{
		underlying: underlying,
		maxRetries: maxRetries,
		sleep:      time.Sleep,
		now:        time.Now,
	}
}

func (c *retrying429Client) Do(req *http.Request) (*http.Response, error) {
	// Buffer the body once so we can replay it on retry.
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		if cerr := req.Body.Close(); cerr != nil {
			return nil, cerr
		}
	}

	var resp *http.Response
	var err error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		resp, err = c.underlying.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		if attempt == c.maxRetries {
			return resp, nil
		}
		// Drain and close the 429 response before retrying.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		wait := c.parseRetryAfter(resp.Header.Get("Retry-After"))
		if wait <= 0 {
			// Exponential backoff: 2s, 4s, 8s, 16s, 32s
			wait = time.Duration(1<<attempt) * 2 * time.Second
		}
		if wait > 60*time.Second {
			wait = 60 * time.Second
		}
		c.sleep(wait)
	}
	return resp, nil
}

func (c *retrying429Client) parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(header); err == nil {
		return t.Sub(c.now())
	}
	return 0
}
