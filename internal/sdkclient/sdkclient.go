package sdkclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hookdeck/hookdeck-go-sdk/client"
	"github.com/hookdeck/hookdeck-go-sdk/option"
)

type Client struct {
	*client.Client

	RawClient *RawClient
}

func InitHookdeckSDKClient(apiBase string, apiKey string, maxAttempts int, providerVersion string) Client {
	if !strings.HasPrefix(apiBase, "http") {
		apiBase = "https://" + apiBase
	}
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)
	hookdeckClient := client.NewClient(
		option.WithBaseURL(apiBase),
		option.WithToken(apiKey),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(uint(maxAttempts)),
	)
	return Client{hookdeckClient, &RawClient{apiKey, apiBase, header}}
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
