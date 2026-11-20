package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteDomain(t *testing.T) {
	tests := []struct {
		name      string
		project   resource
		status    int
		wantCalls int
		wantErr   bool
	}{
		{"outpost project", resource{ID: "tm_1", Type: "outpost"}, 200, 1, false},
		{"outpost project without a domain", resource{ID: "tm_1", Type: "outpost"}, 404, 1, false},
		{"outpost project, server error", resource{ID: "tm_1", Type: "outpost"}, 500, 1, true},
		{"gateway project", resource{ID: "tm_2", Type: "event_gateway"}, 200, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodDelete || r.URL.Path != "/2026-09-01/config/custom_domain" || r.Header.Get(projectHdr) != tc.project.ID {
					t.Errorf("unexpected request %s %s (project %q)", r.Method, r.URL.Path, r.Header.Get(projectHdr))
				}
				w.WriteHeader(tc.status)
			}))
			defer server.Close()
			old := outpostBaseURL
			outpostBaseURL = server.URL
			defer func() { outpostBaseURL = old }()

			err := api{key: "k", client: server.Client()}.deleteDomain(tc.project)
			if (err != nil) != tc.wantErr || calls != tc.wantCalls {
				t.Errorf("err = %v, %d calls; want error %v, %d calls", err, calls, tc.wantErr, tc.wantCalls)
			}
		})
	}
}
