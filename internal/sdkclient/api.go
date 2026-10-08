package sdkclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Do sends a JSON request to path under the API version and decodes the
// response into out. A non-2xx status is returned as *APIError. payload and
// out may be nil.
func (c Client) Do(ctx context.Context, method, path string, payload any, query url.Values, out any) error {
	opts := &RequestOptions{QueryParams: query}
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		opts.Body = bytes.NewReader(data)
		opts.Headers = http.Header{"Content-Type": []string{"application/json"}}
	}
	resp, err := c.RawClient.SendRequest(ctx, method, "/"+APIVersion+path, opts)
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
