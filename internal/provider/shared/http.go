package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// APIError is a non-2xx response.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status %d: %s", e.Status, e.Body)
}

// IsGoneStatus reports whether status means the resource no longer exists.
// Hookdeck answers 404 for unknown ids and 410 for deleted resources.
func IsGoneStatus(status int) bool {
	return status == http.StatusNotFound || status == http.StatusGone
}

// IsNotFound reports whether err is a response for a resource that no
// longer exists.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && IsGoneStatus(apiErr.Status)
}

// Request sends a JSON request and decodes the response into out. A non-2xx
// status is returned as *APIError. out may be nil.
func Request(ctx context.Context, client sdkclient.Client, method, path string, payload any, query url.Values, out any) error {
	opts := &sdkclient.RequestOptions{QueryParams: query}
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		opts.Body = bytes.NewReader(data)
		opts.Headers = http.Header{"Content-Type": []string{"application/json"}}
	}
	resp, err := client.RawClient.SendRequest(ctx, method, "/"+sdkclient.APIVersion+path, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return &APIError{Status: resp.StatusCode, Body: string(body)}
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}

const notFoundSummary = "Resource not found"

// NotFoundDiagnostic marks a Retrieve that got a 404. Resource Read uses
// IsNotFoundDiagnostics to drop the resource from state instead of failing.
func NotFoundDiagnostic(kind, id string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(notFoundSummary, fmt.Sprintf("%s %s no longer exists.", kind, id))
}

// IsNotFoundDiagnostics reports whether diags carries NotFoundDiagnostic.
func IsNotFoundDiagnostics(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if d.Severity() == diag.SeverityError && d.Summary() == notFoundSummary {
			return true
		}
	}
	return false
}
