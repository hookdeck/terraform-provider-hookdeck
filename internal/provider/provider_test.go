package provider_test

import (
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// The types that live in a project, by their name after the provider and
// gateway prefixes.
var projectScopedResources = []string{"connection", "destination", "source", "source_auth", "transformation"}
var projectScopedDataSources = []string{"connection", "destination", "source"}

func registeredResources(t *testing.T) map[string]resource.Resource {
	t.Helper()
	out := map[string]resource.Resource{}
	for _, factory := range provider.New("test")().Resources(t.Context()) {
		r := factory()
		resp := &resource.MetadataResponse{}
		r.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "hookdeck"}, resp)
		out[resp.TypeName] = r
	}
	return out
}

func registeredDataSources(t *testing.T) map[string]datasource.DataSource {
	t.Helper()
	out := map[string]datasource.DataSource{}
	for _, factory := range provider.New("test")().DataSources(t.Context()) {
		d := factory()
		resp := &datasource.MetadataResponse{}
		d.Metadata(t.Context(), datasource.MetadataRequest{ProviderTypeName: "hookdeck"}, resp)
		out[resp.TypeName] = d
	}
	return out
}

func TestResources_gatewayNamesAndDeprecatedAliases(t *testing.T) {
	resources := registeredResources(t)

	for _, name := range projectScopedResources {
		current, legacy := "hookdeck_gateway_"+name, "hookdeck_"+name
		for _, typeName := range []string{current, legacy} {
			r, ok := resources[typeName]
			if !ok {
				t.Errorf("%s is not registered", typeName)
				continue
			}
			resp := &resource.SchemaResponse{}
			r.Schema(t.Context(), resource.SchemaRequest{}, resp)
			description := resp.Schema.Description + resp.Schema.MarkdownDescription

			deprecated := typeName == legacy
			if got := resp.Schema.DeprecationMessage != ""; got != deprecated {
				t.Errorf("%s: deprecation message set = %v, want %v", typeName, got, deprecated)
			}
			if got := strings.Contains(description, "Deprecated"); got != deprecated {
				t.Errorf("%s: description notes deprecation = %v, want %v", typeName, got, deprecated)
			}
			if deprecated && !strings.Contains(resp.Schema.DeprecationMessage, current) {
				t.Errorf("%s: deprecation %q does not name %s", typeName, resp.Schema.DeprecationMessage, current)
			}

			attribute, ok := resp.Schema.Attributes["project_id"]
			if !ok || !attribute.IsOptional() || !attribute.IsComputed() {
				t.Errorf("%s: project_id must be optional and computed", typeName)
			}
			if _, ok := r.(resource.ResourceWithModifyPlan); !ok {
				t.Errorf("%s does not check its project at plan time", typeName)
			}

			mover, ok := r.(resource.ResourceWithMoveState)
			if !ok {
				t.Errorf("%s does not implement MoveState", typeName)
				continue
			}
			if got := len(mover.MoveState(t.Context())) > 0; got == deprecated {
				t.Errorf("%s: accepts moved state = %v, want %v", typeName, got, !deprecated)
			}
		}
	}

	if _, ok := resources["hookdeck_gateway_project"]; !ok {
		t.Error("hookdeck_gateway_project is not registered")
	}
	if _, ok := resources["hookdeck_project"]; ok {
		t.Error("hookdeck_gateway_project has no v2 alias")
	}
}

func TestDataSources_gatewayNamesAndDeprecatedAliases(t *testing.T) {
	dataSources := registeredDataSources(t)

	for _, name := range projectScopedDataSources {
		current, legacy := "hookdeck_gateway_"+name, "hookdeck_"+name
		for _, typeName := range []string{current, legacy} {
			d, ok := dataSources[typeName]
			if !ok {
				t.Errorf("%s is not registered", typeName)
				continue
			}
			resp := &datasource.SchemaResponse{}
			d.Schema(t.Context(), datasource.SchemaRequest{}, resp)

			deprecated := typeName == legacy
			if got := resp.Schema.DeprecationMessage != ""; got != deprecated {
				t.Errorf("%s: deprecation message set = %v, want %v", typeName, got, deprecated)
			}
			if strings.Contains(resp.Schema.DeprecationMessage, "moved block") {
				t.Errorf("%s: a data source cannot be moved: %q", typeName, resp.Schema.DeprecationMessage)
			}
			if got := strings.Contains(resp.Schema.Description, "Deprecated"); got != deprecated {
				t.Errorf("%s: description notes deprecation = %v, want %v", typeName, got, deprecated)
			}
			attribute, ok := resp.Schema.Attributes["project_id"]
			if !ok || !attribute.IsOptional() || !attribute.IsComputed() {
				t.Errorf("%s: project_id must be optional and computed", typeName)
			}
		}
	}
}
