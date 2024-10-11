package sdkclient

import (
	"net/http"

	"github.com/hookdeck/hookdeck-go-sdk/client"
	"github.com/hookdeck/hookdeck-go-sdk/option"
)

func InitHookdeckSDKClient(apiBase string, apiKey string, maxAttempts int, providerVersion string) *client.Client {
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)
	return client.NewClient(
		option.WithBaseURL(apiBase),
		option.WithToken(apiKey),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(uint(maxAttempts)),
	)
}
