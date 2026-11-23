package project

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// validateConfig checks what Terraform needs: a map value cannot be null, and
// a key is in one map only. Refresh gives both maps the one value the API
// holds, so a key in both with different values would show a difference on
// every plan. Which keys exist, and which values are valid, is the API's call.
func validateConfig(config, sensitive map[string]types.String) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, k := range sortedKeys(config) {
		at := path.Root("config").AtMapKey(k)
		if config[k].IsNull() {
			diags.AddAttributeError(at, "Null config value", nullValueMessage(k))
			continue
		}
		if _, ok := sensitive[k]; ok {
			diags.AddAttributeError(at, "Key set twice",
				fmt.Sprintf("%s is set in both config and sensitive_config. Set it in one of them.", k))
		}
	}
	for _, k := range sortedKeys(sensitive) {
		if sensitive[k].IsNull() {
			diags.AddAttributeError(path.Root("sensitive_config").AtMapKey(k), "Null config value", nullValueMessage(k))
		}
	}
	return diags
}

func nullValueMessage(key string) string {
	return fmt.Sprintf("%s is null. Remove the key instead of setting it to null.", key)
}

func planPatch(desired, prior map[string]string, live map[string]*string) map[string]*string {
	patch := map[string]*string{}
	for k, v := range desired {
		current := live[k]
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
