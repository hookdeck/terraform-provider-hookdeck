package project

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const proxyKey = "DESTINATIONS_PROXY_URL"

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func validateConfig(config, sensitive map[string]types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, k := range sortedKeys(config) {
		at := path.Root("config").AtMapKey(k)
		if config[k].IsNull() {
			diags.AddAttributeError(at, "Null config value", nullValueMessage(k))
			continue
		}
		if credentialKeys[k] {
			diags.AddAttributeError(at, "Credential key in config",
				fmt.Sprintf("%s can carry credentials. Set it in sensitive_config instead.", k))
			continue
		}
		if v := config[k]; !v.IsUnknown() && hasUserinfo(v.ValueString()) {
			diags.AddAttributeError(at, "Credentials in config",
				fmt.Sprintf("The value of %s is a URL that carries credentials. Set the key in sensitive_config instead.", k))
			continue
		}
		if _, ok := sensitive[k]; ok {
			diags.AddAttributeError(at, "Key set twice",
				fmt.Sprintf("%s is set in both config and sensitive_config. Set it in one of them.", k))
			continue
		}
		diags.Append(checkKey(at, k)...)
	}
	for _, k := range sortedKeys(sensitive) {
		at := path.Root("sensitive_config").AtMapKey(k)
		if sensitive[k].IsNull() {
			diags.AddAttributeError(at, "Null config value", nullValueMessage(k))
			continue
		}
		diags.Append(checkKey(at, k)...)
	}
	return diags
}

func nullValueMessage(key string) string {
	return fmt.Sprintf("%s is null. Remove the key instead of setting it to null.", key)
}

func checkKey(at path.Path, key string) diag.Diagnostics {
	var diags diag.Diagnostics
	switch reason, rejected := rejectedKeys[key]; {
	case rejected:
		diags.AddAttributeError(at, "Key cannot be managed", fmt.Sprintf("%s cannot be set here: %s.", key, reason))
	case !knownKeys[key]:
		diags.AddAttributeWarning(at, "Unknown config key",
			fmt.Sprintf("%s is not a config key this provider version knows. Check the spelling: Hookdeck rejects a key that does not exist when you apply. If the key was added to Outpost after this provider release, ignore this warning.", key))
	}
	return diags
}

var userinfoPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*://[^/?#]*@`)

// hasUserinfo reports whether value is a URL with a user or password
// (scheme://user:pass@host). A value that does not parse is matched by shape.
func hasUserinfo(value string) bool {
	value = strings.TrimSpace(value)
	u, err := url.Parse(value)
	if err != nil {
		return userinfoPattern.MatchString(value)
	}
	return u.User != nil
}

// sensitiveValue reports whether the data source must return the value in
// sensitive_config although its key is non-sensitive: a URL with userinfo,
// or a value of a URL key that is not a URL with a host.
func sensitiveValue(key, value string) bool {
	if hasUserinfo(value) {
		return true
	}
	if !urlKeys[key] || value == "" {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(value))
	return err != nil || u.Host == ""
}

func planPatch(desired, prior map[string]string, live map[string]*string) map[string]*string {
	patch := map[string]*string{}
	for k, v := range desired {
		current := live[k]
		// The API reads an empty DESTINATIONS_PROXY_URL back as unset, so "" already matches.
		if current == nil && v == "" && k == proxyKey {
			continue
		}
		if current == nil || *current != v {
			value := v
			patch[k] = &value
		}
	}
	for k := range prior {
		if _, kept := desired[k]; !kept && live[k] != nil {
			patch[k] = nil
		}
	}
	return patch
}

func refreshManaged(managed map[string]string, live map[string]*string) map[string]string {
	out := make(map[string]string, len(managed))
	for k, v := range managed {
		switch current := live[k]; {
		case current != nil:
			out[k] = *current
		// "" in state with the key unset live is kept, so the next plan is empty.
		case v == "":
			out[k] = ""
		}
	}
	return out
}

func splitConfig(live map[string]*string) (plain, sensitive map[string]string) {
	plain, sensitive = map[string]string{}, map[string]string{}
	for k, v := range live {
		switch {
		case v == nil:
		case nonSensitiveKeys[k] && !sensitiveValue(k, *v):
			plain[k] = *v
		default:
			sensitive[k] = *v
		}
	}
	return plain, sensitive
}
