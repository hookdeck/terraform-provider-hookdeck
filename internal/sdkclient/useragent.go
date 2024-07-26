package sdkclient

import (
	"encoding/json"
	"net/http"
	"runtime"
)

func initUserAgentHeader(header http.Header, providerVersion string) {
	header.Add("User-Agent", getUserAgent(providerVersion))
	header.Add("X-Hookdeck-Client-User-Agent", getHookdeckUserAgent(providerVersion))
}

func getUserAgent(providerVersion string) string {
	return "Hookdeck/v1 terraform-provider-hookdeck/" + providerVersion
}

// hookdeckClientUserAgent contains information about the current runtime which
// is serialized and sent in the `X-Hookdeck-Client-User-Agent` as additional
// debugging information.
type hookdeckClientUserAgent struct {
	Name      string `json:"name"`
	OS        string `json:"os"`
	Publisher string `json:"publisher"`
	Uname     string `json:"uname"`
	Version   string `json:"version"`
}

func getHookdeckUserAgent(providerVersion string) string {
	hookdeckUserAgent := &hookdeckClientUserAgent{
		Name:      "terraform-provider-hookdeck",
		Version:   providerVersion,
		Publisher: "hookdeck",
		OS:        runtime.GOOS,
		Uname:     getUname(),
	}
	marshaled, err := json.Marshal(hookdeckUserAgent)
	// Encoding this struct should never be a problem, so we're okay to panic
	// in case it is for some reason.
	if err != nil {
		panic(err)
	}

	return string(marshaled)
}
