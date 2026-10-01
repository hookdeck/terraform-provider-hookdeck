package sdkclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/projectscope"
)

type fakeResponse struct {
	status int
	body   string
	err    error
}

// fakeRawClient answers every request with the next queued response, or
// 200 {} when the queue is empty, and records what it was sent.
type fakeRawClient struct {
	responses []fakeResponse
	methods   []string
	paths     []string
	opts      []*RequestOptions
}

func (f *fakeRawClient) SendRequest(_ context.Context, method, path string, opts *RequestOptions) (*http.Response, error) {
	f.methods = append(f.methods, method)
	f.paths = append(f.paths, path)
	f.opts = append(f.opts, opts)
	next := fakeResponse{status: http.StatusOK, body: "{}"}
	if len(f.responses) > 0 {
		next, f.responses = f.responses[0], f.responses[1:]
	}
	if next.err != nil {
		return nil, next.err
	}
	return &http.Response{StatusCode: next.status, Body: io.NopCloser(strings.NewReader(next.body))}, nil
}

func clientWith(kind projectscope.KeyKind, raw RawClientInterface) Client {
	return Client{
		RawClient:     raw,
		Scope:         projectscope.Scope{KeyKind: kind},
		projectChecks: &projectChecks{results: map[string]error{}},
	}
}

func TestWithProject_setsHeaderAndKeepsTheRequest(t *testing.T) {
	raw := &fakeRawClient{}
	client := clientWith(projectscope.KeyKindOrganization, raw).WithProject("tm_1", "")

	body := strings.NewReader(`{"name":"x"}`)
	callerOpts := &RequestOptions{
		Body:    body,
		Headers: http.Header{"Content-Type": []string{"application/json"}},
	}
	if _, err := client.RawClient.SendRequest(t.Context(), "POST", "/sources", callerOpts); err != nil {
		t.Fatal(err)
	}

	sent := raw.opts[0]
	if got := sent.Headers.Get(ProjectHeader); got != "tm_1" {
		t.Errorf("%s = %q, want tm_1", ProjectHeader, got)
	}
	if got := sent.Headers.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, caller headers were dropped", got)
	}
	if sent.Body != body {
		t.Error("body was not forwarded")
	}
	if _, set := callerOpts.Headers[http.CanonicalHeaderKey(ProjectHeader)]; set {
		t.Error("the caller's options were mutated")
	}
}

func TestWithProject_nilOptions(t *testing.T) {
	raw := &fakeRawClient{}
	client := clientWith(projectscope.KeyKindOrganization, raw).WithProject("tm_1", "")
	if _, err := client.RawClient.SendRequest(t.Context(), "GET", "/sources/src_1", nil); err != nil {
		t.Fatal(err)
	}
	if got := raw.opts[0].Headers.Get(ProjectHeader); got != "tm_1" {
		t.Errorf("%s = %q, want tm_1", ProjectHeader, got)
	}
}

func TestUnscopedClient_sendsNoProjectHeader(t *testing.T) {
	raw := &fakeRawClient{}
	client := clientWith(projectscope.KeyKindProject, raw)
	if err := client.Do(t.Context(), "GET", "/sources/src_1", nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := raw.opts[0].Headers.Get(ProjectHeader); got != "" {
		t.Errorf("%s = %q, want none", ProjectHeader, got)
	}
}

func TestWithProject_projectRefusals(t *testing.T) {
	const teamNotFound = `{"code":"NOT_FOUND","status":404,"message":"Team Not Found","data":{"id":"tm_1"}}`
	const sourceNotFound = `{"code":"NOT_FOUND","status":404,"message":"Source Not Found","data":{"id":"src_1"}}`
	const noScope = `{"code":"INSUFFICIENT_SCOPE","status":403,"message":"The API key does not have the required scope for this operation","data":{"required":"gateway.sources.read"}}`

	cases := []struct {
		name    string
		kind    projectscope.KeyKind
		status  int
		body    string
		refused bool
	}{
		{"project key on another project", projectscope.KeyKindProject, 401, "Unauthorized", true},
		{"org key, project in another organization", projectscope.KeyKindOrganization, 401, "Unauthorized", true},
		{"org key, project does not exist", projectscope.KeyKindOrganization, 404, teamNotFound, true},
		{"org key, 404 names the project", projectscope.KeyKindOrganization, 404, `{"message":"Not Found","data":{"id":"tm_1"}}`, true},
		{"org key, no grant on the project", projectscope.KeyKindOrganization, 403, noScope, true},
		{"resource does not exist", projectscope.KeyKindOrganization, 404, sourceNotFound, false},
		{"resource was deleted", projectscope.KeyKindOrganization, 410, `{"code":"GONE"}`, false},
		{"404 without a body", projectscope.KeyKindOrganization, 404, "", false},
		{"project key missing a scope", projectscope.KeyKindProject, 403, noScope, false},
		{"org key, another 403", projectscope.KeyKindOrganization, 403, `{"code":"FORBIDDEN"}`, false},
		{"validation error", projectscope.KeyKindOrganization, 422, `{"code":"UNPROCESSABLE_ENTITY"}`, false},
		{"success", projectscope.KeyKindOrganization, 200, `{"id":"src_1"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := &fakeRawClient{responses: []fakeResponse{{status: tc.status, body: tc.body}}}
			client := clientWith(tc.kind, raw).WithProject("tm_1", "Do this next.")

			resp, err := client.RawClient.SendRequest(t.Context(), "GET", "/sources/src_1", nil)

			var accessErr *ProjectAccessError
			if tc.refused {
				if !errors.As(err, &accessErr) {
					t.Fatalf("err = %v, want *ProjectAccessError", err)
				}
				if accessErr.ProjectID != "tm_1" || accessErr.Status != tc.status {
					t.Errorf("got %+v", accessErr)
				}
				msg := err.Error()
				for _, want := range []string{"cannot access project tm_1", "Do this next.", tc.body} {
					if !strings.Contains(msg, want) {
						t.Errorf("message %q does not contain %q", msg, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want the response", err)
			}
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.status || string(body) != tc.body {
				t.Errorf("response = %d %q, want %d %q", resp.StatusCode, body, tc.status, tc.body)
			}
		})
	}
}

func TestWithProject_transportErrorIsReturned(t *testing.T) {
	boom := errors.New("boom")
	raw := &fakeRawClient{responses: []fakeResponse{{err: boom}}}
	client := clientWith(projectscope.KeyKindOrganization, raw).WithProject("tm_1", "")
	if _, err := client.RawClient.SendRequest(t.Context(), "GET", "/sources/src_1", nil); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}

func TestDo(t *testing.T) {
	raw := &fakeRawClient{responses: []fakeResponse{{status: 200, body: `{"id":"src_1"}`}}}
	client := clientWith(projectscope.KeyKindProject, raw)

	var out struct {
		ID string `json:"id"`
	}
	if err := client.Do(t.Context(), "POST", "/sources", map[string]string{"name": "x"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	if raw.paths[0] != "/"+APIVersion+"/sources" || raw.methods[0] != "POST" {
		t.Errorf("sent %s %s", raw.methods[0], raw.paths[0])
	}
	sent, _ := io.ReadAll(raw.opts[0].Body)
	if string(sent) != `{"name":"x"}` || raw.opts[0].Headers.Get("Content-Type") != "application/json" {
		t.Errorf("payload = %q, headers = %v", sent, raw.opts[0].Headers)
	}
	if out.ID != "src_1" {
		t.Errorf("decoded id = %q", out.ID)
	}
}

func TestDo_nonSuccessIsAPIError(t *testing.T) {
	raw := &fakeRawClient{responses: []fakeResponse{{status: 410, body: "gone"}}}
	err := clientWith(projectscope.KeyKindProject, raw).Do(t.Context(), "GET", "/sources/src_1", nil, nil, nil)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 410 || apiErr.Body != "gone" {
		t.Fatalf("err = %#v", err)
	}
	if !IsNotFound(err) {
		t.Error("410 should count as not found")
	}
	if IsNotFound(&APIError{Status: 500}) || IsNotFound(errors.New("x")) {
		t.Error("only 404 and 410 count as not found")
	}
}

func TestKeyProjectID(t *testing.T) {
	boom := errors.New("boom")
	cases := []struct {
		name       string
		response   fakeResponse
		want       string
		wantStatus int
		wantErr    bool
	}{
		{"one project", fakeResponse{status: 200, body: `[{"id":"tm_key"}]`}, "tm_key", 0, false},
		{"no project", fakeResponse{status: 200, body: `[]`}, "", 0, true},
		{"several projects", fakeResponse{status: 200, body: `[{"id":"tm_a"},{"id":"tm_b"}]`}, "", 0, true},
		{"project without id", fakeResponse{status: 200, body: `[{}]`}, "", 0, true},
		{"not allowed to list projects", fakeResponse{status: 403, body: "{}"}, "", 403, true},
		{"invalid key", fakeResponse{status: 401, body: "Unauthorized"}, "", 401, true},
		{"transport error", fakeResponse{err: boom}, "", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := &fakeRawClient{responses: []fakeResponse{tc.response}}
			got, err := clientWith(projectscope.KeyKindProject, raw).KeyProjectID(t.Context())
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Fatalf("got (%q, %v), want (%q, err=%v)", got, err, tc.want, tc.wantErr)
			}
			var apiErr *APIError
			if tc.wantStatus != 0 && (!errors.As(err, &apiErr) || apiErr.Status != tc.wantStatus) {
				t.Fatalf("err = %#v, want status %d", err, tc.wantStatus)
			}
			if raw.paths[0] != "/"+APIVersion+"/projects" {
				t.Errorf("path = %s", raw.paths[0])
			}
			if got := raw.opts[0].Headers.Get(ProjectHeader); got != "" {
				t.Errorf("lookup sent %s = %q", ProjectHeader, got)
			}
		})
	}
}

func TestCheckProject(t *testing.T) {
	boom := errors.New("boom")
	raw := &fakeRawClient{responses: []fakeResponse{
		{status: 200, body: `{"id":"tm_ok"}`},
		{status: 404, body: `{"message":"Not Found"}`},
		{status: 403, body: `{"code":"INSUFFICIENT_SCOPE"}`},
		{err: boom},
		{status: 200, body: `{"id":"tm_flaky"}`},
	}}
	client := clientWith(projectscope.KeyKindOrganization, raw)
	ctx := t.Context()

	if err := client.CheckProject(ctx, "tm_ok", ""); err != nil {
		t.Errorf("visible project: %v", err)
	}
	var accessErr *ProjectAccessError
	if err := client.CheckProject(ctx, "tm_missing", "hint"); !errors.As(err, &accessErr) || accessErr.ProjectID != "tm_missing" || accessErr.Hint != "hint" {
		t.Errorf("missing project: %#v", err)
	}
	if err := client.CheckProject(ctx, "tm_unreadable", ""); err != nil {
		t.Errorf("a key that may not read projects cannot tell: %v", err)
	}
	if err := client.CheckProject(ctx, "tm_flaky", ""); !errors.Is(err, boom) {
		t.Errorf("transport error: %v", err)
	}

	// Answers are kept; a failed check is asked again.
	if err := client.CheckProject(ctx, "tm_ok", ""); err != nil {
		t.Error(err)
	}
	if err := client.CheckProject(ctx, "tm_missing", ""); !errors.As(err, &accessErr) {
		t.Error(err)
	}
	if err := client.CheckProject(ctx, "tm_flaky", ""); err != nil {
		t.Errorf("retry after a transport error: %v", err)
	}
	if len(raw.paths) != 5 {
		t.Errorf("%d requests, want 5: %v", len(raw.paths), raw.paths)
	}
}
