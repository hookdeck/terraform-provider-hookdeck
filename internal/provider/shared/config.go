package shared

import (
	"context"
	"fmt"
	"strings"

	"terraform-provider-hookdeck/internal/configjson"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configKeysPrivateKey is the private state key that records which config keys
// the provider has sent to Hookdeck. State holds Hookdeck's full config, so
// without it a key removed from the configuration could not be told apart
// from a key Hookdeck added.
const configKeysPrivateKey = "config_keys"

const configDescription = "Only the keys you set are compared with Hookdeck; a change made outside Terraform to " +
	"one of them shows in the plan, and the next apply restores your value. Other keys, including defaults Hookdeck " +
	"adds, are kept in state and not compared. Removing a key clears it in Hookdeck. Holds the secrets Hookdeck " +
	"returns, such as authentication tokens. See [Source and destination config]" +
	"(https://registry.terraform.io/providers/hookdeck/hookdeck/latest/docs/guides/config)."

// ConfigAttribute returns the config attribute of a source or destination.
// lead is the first sentence of its description and kind ("source" or
// "destination") is used in plan warnings.
func ConfigAttribute(lead, kind string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Sensitive:           true,
		MarkdownDescription: lead + " " + configDescription,
		PlanModifiers:       []planmodifier.String{ConfigPlanModifier(kind)},
		CustomType:          jsontypes.NormalizedType{},
	}
}

// ConfigDataSourceAttribute returns the config attribute of a source or
// destination data source.
func ConfigDataSourceAttribute(lead string) datasourceschema.StringAttribute {
	return datasourceschema.StringAttribute{
		Computed:            true,
		Sensitive:           true,
		MarkdownDescription: strings.TrimSuffix(lead, ".") + ", as Hookdeck returns it, including defaults Hookdeck adds and authentication secrets.",
		CustomType:          jsontypes.NormalizedType{},
	}
}

// ConfigPlanModifier plans the config attribute. config is Computed, so state
// holds everything Hookdeck returns while the configuration holds only what
// the user set. The modifier keeps the planned value equal to state when
// every key in the configuration already holds the configured value, so the
// plan stays empty, and plans the configuration otherwise.
//
// This cannot be done with semantic equality: the framework does not run it
// at plan time, so it cannot hide a difference in a subset of the keys.
func ConfigPlanModifier(kind string) planmodifier.String {
	return configPlanModifier{kind: kind}
}

type configPlanModifier struct{ kind string }

func (m configPlanModifier) Description(context.Context) string {
	return "Compares the configured keys with the configuration Hookdeck returns."
}

func (m configPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m configPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() {
		return
	}
	managed, recorded, diags := ManagedConfigPaths(ctx, req.Private, req.StateValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	planned, missing, err := planConfig(req.ConfigValue, req.StateValue, req.PlanValue, managed, recorded)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid config", err.Error())
		return
	}
	for _, key := range missing {
		resp.Diagnostics.AddAttributeWarning(req.Path, "Config key not returned by Hookdeck",
			fmt.Sprintf("Hookdeck does not return config key `%s` for this %s, so it has no effect. Check the key name or remove it from `config`.", key, m.kind))
	}
	resp.PlanValue = planned
}

// planConfig decides the planned value of config. plan is the value the
// framework proposed: the configuration when set, otherwise state, or unknown
// on create. managed holds the key paths the provider sent earlier; recorded
// is false when they are not known (imported and not yet applied, or private
// data that cannot be read), in which case nothing counts as removed or
// added. missing lists configured keys Hookdeck does not return.
//
// After an import (recorded false) a configuration that matches Hookdeck plans
// no change, so its keys are recorded by the next apply that changes config.
func planConfig(config, state, plan types.String, managed configjson.Paths, recorded bool) (planned types.String, missing []string, err error) {
	if state.IsNull() || state.IsUnknown() || config.IsUnknown() {
		return plan, nil, nil
	}
	if !recorded {
		managed = nil
	}
	if config.IsNull() {
		// The configuration has no config at all: everything the provider
		// sent is cleared, and the new value is not known until Hookdeck answers.
		if len(managed) > 0 {
			return types.StringUnknown(), nil, nil
		}
		return state, nil, nil
	}

	user, err := configjson.Parse(config.ValueString())
	if err != nil {
		return plan, nil, err
	}
	current, err := configjson.Parse(state.ValueString())
	if err != nil {
		// State is not valid JSON: plan the configuration to repair it.
		return config, nil, nil //nolint:nilerr // not an error for the caller, the plan repairs it
	}
	covers, missing := configjson.Covers(user, current, managed)
	if !covers || len(configjson.Removed(managed, user)) > 0 {
		return config, missing, nil
	}
	// A key the provider has not sent before is recorded by the apply, even
	// when Hookdeck already holds its value: without it, removing the key
	// later would change nothing.
	if recorded && len(configjson.Added(managed, user)) > 0 {
		return config, missing, nil
	}
	return state, missing, nil
}

type privateGetter interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

type privateSetter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

type configKeys int

const (
	// configKeysAbsent: nothing stored. State written before the keys were
	// tracked, which holds exactly the configuration the user applied (none
	// when config was not set). Private data is lost only by editing state by
	// hand; such a state's full config then counts as set.
	configKeysAbsent configKeys = iota
	// configKeysUnrecorded: the provider never sent keys for this resource
	// (imported), or the stored value cannot be read. Which keys the user
	// manages is not known, and state cannot tell: it holds Hookdeck's full config.
	configKeysUnrecorded
	configKeysRecorded
)

func readConfigKeys(ctx context.Context, p privateGetter) (configjson.Paths, configKeys, diag.Diagnostics) {
	if p == nil {
		return nil, configKeysAbsent, nil
	}
	raw, diags := p.GetKey(ctx, configKeysPrivateKey)
	if diags.HasError() {
		return nil, configKeysUnrecorded, diags
	}
	if len(raw) == 0 {
		return nil, configKeysAbsent, diags
	}
	paths, recorded, err := configjson.UnmarshalPaths(raw)
	if err != nil || !recorded {
		return nil, configKeysUnrecorded, diags
	}
	return paths, configKeysRecorded, diags
}

// ManagedConfigPaths returns the config key paths the provider sent to
// Hookdeck. recorded is false when they are not known: an imported resource
// that was not applied yet, or private data that cannot be read.
//
// When nothing is stored the resource comes from a version that did not track
// them, and state still holds exactly the configuration the user applied, so
// the paths are taken from it; without a config in state none were set.
func ManagedConfigPaths(ctx context.Context, p privateGetter, state types.String) (paths configjson.Paths, recorded bool, diags diag.Diagnostics) {
	paths, kind, diags := readConfigKeys(ctx, p)
	switch kind {
	case configKeysRecorded:
		return paths, true, diags
	case configKeysAbsent:
		if state.IsUnknown() {
			return nil, false, diags
		}
		if state.IsNull() {
			return configjson.Paths{}, true, diags
		}
		user, err := configjson.Parse(state.ValueString())
		if err != nil {
			return nil, false, diags
		}
		return configjson.PathsOf(user), true, diags
	}
	return nil, false, diags
}

// SetManagedConfigPaths records the key paths of the configured config. Only
// key names are stored, never values.
func SetManagedConfigPaths(ctx context.Context, p privateSetter, config jsontypes.Normalized) diag.Diagnostics {
	var diags diag.Diagnostics
	paths := configjson.Paths{}
	if !config.IsNull() && !config.IsUnknown() {
		user, err := configjson.Parse(config.ValueString())
		if err != nil {
			diags.AddError("Invalid config", err.Error())
			return diags
		}
		paths = configjson.PathsOf(user)
	}
	raw, err := paths.Marshal()
	if err != nil {
		diags.AddError("Error recording config keys", err.Error())
		return diags
	}
	return p.SetKey(ctx, configKeysPrivateKey, raw)
}

// SeedManagedConfigPaths stores what Read can tell about the keys the
// provider manages, when nothing is stored yet. State from a version that did
// not track them holds exactly the user's configuration: its keys are the
// managed ones, and none when it has no config. An imported resource is marked
// as not recorded by MarkConfigUnrecorded, and later reads leave that alone:
// state then holds Hookdeck's full config, which says nothing about what the
// user manages. The first apply that changes config records them.
//
// Private data that exists but cannot be read is never replaced from state.
func SeedManagedConfigPaths(ctx context.Context, p interface {
	privateGetter
	privateSetter
}, state jsontypes.Normalized) diag.Diagnostics {
	_, kind, diags := readConfigKeys(ctx, p)
	if diags.HasError() || kind != configKeysAbsent {
		return diags
	}
	diags.Append(SetManagedConfigPaths(ctx, p, state)...)
	return diags
}

// MarkConfigUnrecorded marks a resource being imported as having no recorded
// config keys. Call it from ImportState after the id is set; the marker
// reaches the Read that follows.
func MarkConfigUnrecorded(ctx context.Context, resp *resource.ImportStateResponse) {
	if resp.Diagnostics.HasError() {
		return
	}
	raw, err := configjson.MarshalUnrecorded()
	if err != nil {
		resp.Diagnostics.AddError("Error recording config keys", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, configKeysPrivateKey, raw)...)
}

// ConfigFromAPI returns the config of an API response as canonical JSON, with
// {} when the response has none.
func ConfigFromAPI(obj map[string]any) (jsontypes.Normalized, diag.Diagnostics) {
	var diags diag.Diagnostics
	config, _ := obj["config"].(map[string]any)
	s, err := configjson.Canonical(config)
	if err != nil {
		diags.AddError("Error parsing config", err.Error())
		return jsontypes.NewNormalizedNull(), diags
	}
	return jsontypes.NewNormalizedValue(s), diags
}

// DecodeObject decodes an API response body, keeping numbers exact.
func DecodeObject(body []byte) (map[string]any, error) {
	obj, err := configjson.Parse(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ConfigUnchanged reports whether an update leaves config as it is: the
// planned config equals the stored one and the type is the same. The update
// then sends no config, so a configuration the plan ignored (ignore_changes)
// is not applied on the side.
func ConfigUnchanged(plan, state jsontypes.Normalized, planType, stateType types.String) bool {
	return !plan.IsUnknown() && !state.IsUnknown() && plan.Equal(state) && planType.Equal(stateType)
}
