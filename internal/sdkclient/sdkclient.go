package sdkclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	RawClient *RawClient
}

const (
	defaultAPIBase = "api.hookdeck.com"
)

func InitHookdeckSDKClient(apiBase string, apiKey string, providerVersion string) Client {
	if apiBase == "" {
		apiBase = defaultAPIBase
	}
	if !strings.HasPrefix(apiBase, "http") {
		apiBase = "https://" + apiBase
	}
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)

	return Client{&RawClient{apiKey, apiBase, header}}
}

type RawClient struct {
	apiKey  string
	apiBase string
	header  http.Header
}

type RequestOptions struct {
	Body        io.Reader
	Headers     http.Header
	QueryParams url.Values
}

func (c *RawClient) SendRequest(method, path string, opts *RequestOptions) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.apiBase, path)
	if opts != nil && opts.QueryParams != nil {
		url += "?" + opts.QueryParams.Encode()
	}

	req, err := http.NewRequest(method, url, opts.Body)
	if err != nil {
		return nil, err
	}

	// Set default headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	for key, value := range c.header {
		for _, v := range value {
			req.Header.Add(key, v)
		}
	}

	// Set custom headers if provided
	if opts != nil && opts.Headers != nil {
		for key, values := range opts.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	return http.DefaultClient.Do(req)
}
