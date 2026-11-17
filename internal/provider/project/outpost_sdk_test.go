package project

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"testing"
	"time"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/sdkclient"
)

type queuedAPI struct {
	responses []response
	calls     []string
	bodies    []string
}

func (q *queuedAPI) SendRequest(_ context.Context, method, path string, opts *sdkclient.RequestOptions) (*http.Response, error) {
	q.calls = append(q.calls, method+" "+strings.TrimPrefix(path, "/"+sdkclient.APIVersion))
	body := ""
	if opts != nil && opts.Body != nil {
		b, _ := io.ReadAll(opts.Body)
		body = string(b)
	}
	q.bodies = append(q.bodies, body)
	if len(q.responses) == 0 {
		return nil, fmt.Errorf("unexpected request %s %s", method, path)
	}
	next := q.responses[0]
	if len(q.responses) > 1 {
		q.responses = q.responses[1:]
	}
	return &http.Response{StatusCode: next.status, Body: io.NopCloser(strings.NewReader(next.body))}, nil
}

func fastPolling(t *testing.T) {
	t.Helper()
	oldInterval, oldMax, oldLock := pollInterval, pollMaxInterval, lockRetryDelay
	pollInterval, pollMaxInterval, lockRetryDelay = time.Millisecond, 4*time.Millisecond, time.Millisecond
	t.Cleanup(func() { pollInterval, pollMaxInterval, lockRetryDelay = oldInterval, oldMax, oldLock })
}

const (
	notProvisioned = `{"code":"OUTPOST_NOT_PROVISIONED","status":409,"message":"not provisioned"}`
	configLocked   = `{"code":"OUTPOST_CONFIG_LOCKED","status":409,"message":"locked"}`
)

func TestWaitProvisioned(t *testing.T) {
	tests := []struct {
		name      string
		responses []response
		timeout   time.Duration
		wantCalls int
		wantError string
	}{
		{name: "not provisioned three times, then ready", responses: []response{{409, notProvisioned}, {409, notProvisioned}, {409, notProvisioned}, {200, `{}`}}, timeout: time.Second, wantCalls: 4},
		{name: "server error, then ready", responses: []response{{503, `unavailable`}, {200, `{}`}}, timeout: time.Second, wantCalls: 2},
		{name: "rate limited, then ready", responses: []response{{429, `slow down`}, {200, `{}`}}, timeout: time.Second, wantCalls: 2},
		{name: "404 fails at once", responses: []response{{404, `{"code":"NOT_FOUND"}`}}, timeout: time.Second, wantCalls: 1, wantError: "status 404"},
		{name: "team not found fails at once", responses: []response{{404, `{"message":"Team Not Found"}`}}, timeout: time.Second, wantCalls: 1, wantError: "cannot access project tm_1"},
		{name: "401 fails at once", responses: []response{{401, `{"message":"Unauthorized"}`}}, timeout: time.Second, wantCalls: 1, wantError: "cannot access project tm_1"},
		{name: "403 scope fails at once", responses: []response{{403, `{"code":"INSUFFICIENT_SCOPE"}`}}, timeout: time.Second, wantCalls: 1, wantError: "INSUFFICIENT_SCOPE"},
		{name: "400 fails at once", responses: []response{{400, `{"code":"BAD"}`}}, timeout: time.Second, wantCalls: 1, wantError: "status 400"},
		{name: "invalid body fails at once", responses: []response{{200, `not json`}}, timeout: time.Second, wantCalls: 1, wantError: "invalid character"},
		{name: "other 409 fails at once", responses: []response{{409, configLocked}}, timeout: time.Second, wantCalls: 1, wantError: "OUTPOST_CONFIG_LOCKED"},
		{name: "deadline", responses: []response{{409, notProvisioned}}, timeout: 30 * time.Millisecond, wantCalls: -1, wantError: "Outpost project tm_1 was created but is not ready after 30ms"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fastPolling(t)
			api := &queuedAPI{responses: tc.responses}
			ctx, cancel := context.WithTimeout(t.Context(), tc.timeout)
			defer cancel()
			err := waitProvisioned(ctx, outpostAPI(sdkclient.Client{OutpostRawClient: api, Scope: projectscope.Scope{KeyKind: projectscope.KeyKindOrganization}}, "tm_1"), "tm_1", tc.timeout)
			if tc.wantError == "" && err != nil {
				t.Fatal(err)
			}
			if tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.wantError)
			}
			if tc.wantCalls >= 0 && len(api.calls) != tc.wantCalls {
				t.Errorf("calls = %v, want %d", api.calls, tc.wantCalls)
			}
			for _, call := range api.calls {
				if call != "GET /config" {
					t.Errorf("unexpected call %s", call)
				}
			}
		})
	}
}

func TestPatchConfig(t *testing.T) {
	t.Run("lock, then ok", func(t *testing.T) {
		fastPolling(t)
		api := &queuedAPI{responses: []response{{409, configLocked}, {200, `{"TOPICS":"a"}`}}}
		got, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": ptr("a")})
		if err != nil {
			t.Fatal(err)
		}
		if got["TOPICS"] == nil || *got["TOPICS"] != "a" || len(api.calls) != 2 {
			t.Errorf("config = %v, calls = %v", got, api.calls)
		}
		for _, body := range api.bodies {
			if body != `{"TOPICS":"a"}` {
				t.Errorf("body = %s", body)
			}
		}
	})

	t.Run("gives up after the retries", func(t *testing.T) {
		fastPolling(t)
		api := &queuedAPI{responses: []response{{409, configLocked}}}
		_, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": nil})
		if err == nil || len(api.calls) != lockRetries+1 {
			t.Errorf("error = %v, calls = %d, want %d", err, len(api.calls), lockRetries+1)
		}
	})

	t.Run("5xx and 429, then ok", func(t *testing.T) {
		fastPolling(t)
		api := &queuedAPI{responses: []response{{503, `{}`}, {429, `{}`}, {200, `{"TOPICS":"a"}`}}}
		got, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": ptr("a")})
		if err != nil || got["TOPICS"] == nil || len(api.calls) != 3 {
			t.Errorf("config = %v, error = %v, calls = %v", got, err, api.calls)
		}
	})

	t.Run("gives up on a persistent 5xx", func(t *testing.T) {
		fastPolling(t)
		api := &queuedAPI{responses: []response{{500, `{}`}}}
		_, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": nil})
		if err == nil || len(api.calls) != transientRetries+1 {
			t.Errorf("error = %v, calls = %d, want %d", err, len(api.calls), transientRetries+1)
		}
	})

	t.Run("a 4xx is not retried", func(t *testing.T) {
		fastPolling(t)
		api := &queuedAPI{responses: []response{{403, `{}`}}}
		_, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": nil})
		if err == nil || len(api.calls) != 1 {
			t.Errorf("error = %v, calls = %v", err, api.calls)
		}
	})

	t.Run("422 carries the API messages", func(t *testing.T) {
		api := &queuedAPI{responses: []response{{422, `{"code":"VALIDATION","message":"\"NOPE\" is not allowed"}`}}}
		_, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"NOPE": ptr("x")})
		if err == nil || !strings.Contains(err.Error(), `\"NOPE\" is not allowed`) || len(api.calls) != 1 {
			t.Errorf("error = %v, calls = %v", err, api.calls)
		}
	})

	t.Run("null removes a key", func(t *testing.T) {
		api := &queuedAPI{responses: []response{{200, `{}`}}}
		if _, err := patchConfig(t.Context(), sdkclient.Client{RawClient: api}, map[string]*string{"TOPICS": nil}); err != nil {
			t.Fatal(err)
		}
		if api.bodies[0] != `{"TOPICS":null}` {
			t.Errorf("body = %s", api.bodies[0])
		}
	})
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestRetryableWait(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"connection refused", &url.Error{Op: "Get", URL: "https://x", Err: syscall.ECONNREFUSED}, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"timeout", &url.Error{Op: "Get", URL: "https://x", Err: timeoutError{}}, true},
		{"connection reset", &url.Error{Op: "Get", URL: "https://x", Err: syscall.ECONNRESET}, true},
		{"EOF", &url.Error{Op: "Get", URL: "https://x", Err: io.EOF}, true},
		{"DNS name not found", &url.Error{Op: "Get", URL: "https://x", Err: &net.DNSError{Err: "no such host", Name: "x", IsNotFound: true}}, false},
		{"DNS timeout", &url.Error{Op: "Get", URL: "https://x", Err: &net.DNSError{Err: "i/o timeout", Name: "x", IsTimeout: true}}, true},
		{"TLS certificate", &url.Error{Op: "Get", URL: "https://x", Err: x509.UnknownAuthorityError{}}, false},
		{"TLS certificate invalid", &url.Error{Op: "Get", URL: "https://x", Err: x509.CertificateInvalidError{Reason: x509.Expired}}, false},
		{"access refused", &sdkclient.ProjectAccessError{ProjectID: "tm_1", Status: 401}, false},
		{"other error", fmt.Errorf("boom"), false},
	}
	for _, tc := range tests {
		if got := retryableWait(tc.err); got != tc.want {
			t.Errorf("%s: retryable = %v, want %v", tc.name, got, tc.want)
		}
	}
}
