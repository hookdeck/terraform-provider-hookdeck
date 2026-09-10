package destination

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// API version upgrade: 2025-07-01 -> 2026-09-01
//
// Everything specific to this upgrade lives in this file. ValidateConfig in
// resource.go calls validateConfigForAPIVersion20260901; drop that call and this file
// once the migration window is over. A future version bump gets its own file
// and function, named after the API version, alongside this one.
//
// 2026-09-01 replaces three top-level destination config fields with a single
// `delivery_policy` object. The API strips unknown fields rather than
// rejecting them, so a config still using the old names would apply without
// error while the API silently kept whatever policy it already had. Rejecting
// the old names at plan time makes the upgrade loud instead of silent.

// removedConfigField20260901 maps a config field the API no longer accepts to the
// field that replaced it.
type removedConfigField20260901 struct {
	name, replacement string
}

var removedConfigFields20260901 = []removedConfigField20260901{
	{"rate_limit", "delivery_policy.rate"},
	{"rate_limit_period", "delivery_policy.period"},
	{"delivery_groups", "delivery_policy.groups"},
}

// removedConfigFieldsIn20260901 returns the removed field names present at the top
// level of a destination config JSON document, in removedConfigFields20260901 order.
// Malformed JSON yields nothing; the API reports it on apply.
func removedConfigFieldsIn20260901(config string) []string {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		return nil
	}
	var found []string
	for _, field := range removedConfigFields20260901 {
		if _, ok := parsed[field.name]; ok {
			found = append(found, field.name)
		}
	}
	return found
}

// validateConfigForAPIVersion20260901 returns an error diagnostic on `config` when it
// still uses fields removed by the current API version.
func validateConfigForAPIVersion20260901(config jsontypes.Normalized) diag.Diagnostics {
	var diags diag.Diagnostics
	if config.IsUnknown() || config.IsNull() {
		return diags
	}
	removed := removedConfigFieldsIn20260901(config.ValueString())
	if len(removed) == 0 {
		return diags
	}
	mapping := make([]string, 0, len(removedConfigFields20260901))
	for _, field := range removedConfigFields20260901 {
		mapping = append(mapping, field.name+" -> "+field.replacement)
	}
	diags.AddAttributeError(
		path.Root("config"),
		"Destination config uses removed fields",
		fmt.Sprintf(
			"%s: not supported by Hookdeck API version %s. Move these settings under \"delivery_policy\" (%s). See the v2.3 to v2.4 migration guide.",
			strings.Join(removed, ", "), apiVersion, strings.Join(mapping, ", "),
		),
	)
	return diags
}
