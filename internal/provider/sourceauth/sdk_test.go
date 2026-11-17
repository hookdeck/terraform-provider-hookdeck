package sourceauth

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type scriptedClient struct {
	// status is the answer per method.
	status map[string]int
	calls  []string
}

func (s *scriptedClient) SendRequest(_ context.Context, method, path string, _ *sdkclient.RequestOptions) (*http.Response, error) {
	s.calls = append(s.calls, method)
	status, ok := s.status[method]
	if !ok {
		status = http.StatusOK
	}
	body := `{"id":"src_1","type":"HTTP","config":{"auth_type":"API_KEY","auth":{}}}`
	if status > 299 {
		body = `{"message":"gone"}`
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		status    map[string]int
		wantErr   bool
		wantCalls []string
	}{
		{"deleted", nil, false, []string{"GET", "PUT"}},
		{"source gone on read, 404", map[string]int{"GET": 404}, false, []string{"GET"}},
		{"source gone on read, 410", map[string]int{"GET": 410}, false, []string{"GET"}},
		{"source gone on write, 404", map[string]int{"PUT": 404}, false, []string{"GET", "PUT"}},
		{"source gone on write, 410", map[string]int{"PUT": 410}, false, []string{"GET", "PUT"}},
		{"read fails", map[string]int{"GET": 500}, true, []string{"GET"}},
		{"write fails", map[string]int{"PUT": 500}, true, []string{"GET", "PUT"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw := &scriptedClient{status: tc.status}
			m := &sourceAuthResourceModel{SourceID: types.StringValue("src_1")}
			diags := m.Delete(t.Context(), &sdkclient.Client{RawClient: raw})
			if diags.HasError() != tc.wantErr {
				t.Errorf("errors = %v, want error %v", diags, tc.wantErr)
			}
			if strings.Join(raw.calls, " ") != strings.Join(tc.wantCalls, " ") {
				t.Errorf("calls = %v, want %v", raw.calls, tc.wantCalls)
			}
		})
	}
}
