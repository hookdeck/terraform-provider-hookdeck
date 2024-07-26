package sdkclient

import (
	"net/http"

	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

func InitHookdeckSDKClient(apiBase string, apiKey string, providerVersion string) *hookdeckClient.Client {
	header := http.Header{}
	initUserAgentHeader(header, providerVersion)
	return hookdeckClient.NewClient(
		hookdeckClient.WithBaseURL(apiBase),
		hookdeckClient.WithAuthToken(apiKey),
		hookdeckClient.WithHTTPHeader(header),
	)
}
