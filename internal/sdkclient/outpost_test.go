package sdkclient

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type seenRequest struct {
	server, method, path, project, auth string
}

func recordingServer(t *testing.T, name string, mu *sync.Mutex, seen *[]seenRequest) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*seen = append(*seen, seenRequest{name, r.Method, r.URL.Path, r.Header.Get(ProjectHeader), r.Header.Get("Authorization")})
		mu.Unlock()
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestOutpost_usesTheOutpostHostWithVersionAndProjectHeader(t *testing.T) {
	var mu sync.Mutex
	var seen []seenRequest
	main := recordingServer(t, "main", &mu, &seen)
	outpost := recordingServer(t, "outpost", &mu, &seen)
	client := InitHookdeckSDKClient(main.URL, outpost.URL, "hd_org_key", "test")

	if err := client.Outpost().WithProject("tm_1", "").Do(t.Context(), http.MethodGet, "/config", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := client.Do(t.Context(), http.MethodGet, "/projects", nil, nil, nil); err != nil {
		t.Fatal(err)
	}

	want := []seenRequest{
		{"outpost", "GET", "/" + APIVersion + "/config", "tm_1", "Bearer hd_org_key"},
		{"main", "GET", "/" + APIVersion + "/projects", "", "Bearer hd_org_key"},
	}
	if len(seen) != len(want) {
		t.Fatalf("requests = %+v, want %+v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("request %d = %+v, want %+v", i, seen[i], want[i])
		}
	}
}

func TestOutpost_sharesTheRateLimiter(t *testing.T) {
	limiter := &MockRateLimiter{}
	var mu sync.Mutex
	var seen []seenRequest
	main := recordingServer(t, "main", &mu, &seen)
	outpost := recordingServer(t, "outpost", &mu, &seen)
	client := InitHookdeckSDKClient(main.URL, outpost.URL, "hd_org_key", "test", WithRateLimiter(limiter))

	_ = client.Do(t.Context(), http.MethodGet, "/projects", nil, nil, nil)
	_ = client.Outpost().Do(t.Context(), http.MethodGet, "/config", nil, nil, nil)

	if got := limiter.GetWaitCalls(); got != 2 {
		t.Errorf("limiter waited %d times for 2 requests across both hosts, want 2", got)
	}
}

func TestOutpost_defaultHostAndFallback(t *testing.T) {
	client := InitHookdeckSDKClient("", "", "hd_org_key", "test")
	outpost, _ := client.OutpostRawClient.(*RawClient)
	main, _ := client.RawClient.(*RawClient)
	if outpost == nil || outpost.apiBase != "https://api.outpost.hookdeck.com" {
		t.Errorf("default Outpost base = %+v", outpost)
	}
	if main == nil || main.apiBase != "https://api.hookdeck.com" {
		t.Errorf("default base = %+v", main)
	}

	raw := &fakeRawClient{}
	fake := Client{RawClient: raw}
	if fake.Outpost().RawClient != RawClientInterface(raw) {
		t.Error("Outpost() on a client without an Outpost raw client must keep RawClient")
	}
}

func TestAPIError_codeAndIsCode(t *testing.T) {
	err := &APIError{Status: 409, Body: `{"code":"OUTPOST_NOT_PROVISIONED","message":"x"}`}
	if err.Code() != "OUTPOST_NOT_PROVISIONED" {
		t.Errorf("Code() = %q", err.Code())
	}
	if !IsCode(err, 409, "OUTPOST_NOT_PROVISIONED") {
		t.Error("IsCode did not match")
	}
	if IsCode(err, 404, "OUTPOST_NOT_PROVISIONED") || IsCode(err, 409, "OTHER") || IsCode(nil, 409, "OUTPOST_NOT_PROVISIONED") {
		t.Error("IsCode matched too much")
	}
	if (&APIError{Status: 500, Body: "not json"}).Code() != "" {
		t.Error("Code() of a non-JSON body must be empty")
	}
}
