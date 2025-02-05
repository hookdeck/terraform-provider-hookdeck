package sdkclient

import (
	"net/http"

	"github.com/hookdeck/hookdeck-go-sdk/client"
	"github.com/hookdeck/hookdeck-go-sdk/option"
)

type Client struct {
	*client.Client
}

func InitHookdeckSDKClient(apiBase string, apiKey string, maxAttempts int, providerVersion string) Client {
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)
	hookdeckClient := client.NewClient(
		option.WithBaseURL(apiBase),
		option.WithToken(apiKey),
		option.WithHTTPHeader(header),
		option.WithMaxAttempts(uint(maxAttempts)),
	)
	return Client{hookdeckClient}
}
