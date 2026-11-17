package provider_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-hookdeck/internal/provider"
	"terraform-provider-hookdeck/internal/sdkclient"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func configuredClient(t *testing.T, outpostAPIBase *string) sdkclient.Client {
	t.Helper()
	p := provider.New("test")()
	schemaResp := &fwprovider.SchemaResponse{}
	p.Schema(t.Context(), fwprovider.SchemaRequest{}, schemaResp)

	attrTypes := map[string]tftypes.Type{}
	values := map[string]tftypes.Value{}
	for name := range schemaResp.Schema.Attributes {
		attrTypes[name] = tftypes.String
		values[name] = tftypes.NewValue(tftypes.String, nil)
	}
	values["api_key"] = tftypes.NewValue(tftypes.String, "hd_org_test")
	if outpostAPIBase != nil {
		values["outpost_api_base"] = tftypes.NewValue(tftypes.String, *outpostAPIBase)
	}
	raw := tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypes}, values)

	resp := &fwprovider.ConfigureResponse{}
	p.Configure(t.Context(), fwprovider.ConfigureRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	client, ok := resp.ResourceData.(sdkclient.Client)
	if !ok {
		t.Fatalf("resource data is %T", resp.ResourceData)
	}
	return client
}

func countingServer(t *testing.T) (*httptest.Server, *int) {
	t.Helper()
	var hits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	return server, &hits
}

func TestProviderConfig_outpostAPIBase(t *testing.T) {
	t.Run("environment variable", func(t *testing.T) {
		server, hits := countingServer(t)
		t.Setenv("HOOKDECK_OUTPOST_API_BASE", server.URL)

		client := configuredClient(t, nil)
		if err := client.Outpost().Do(t.Context(), http.MethodGet, "/config", nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if *hits != 1 {
			t.Errorf("Outpost API requests = %d, want 1", *hits)
		}
	})

	t.Run("attribute wins over the environment variable", func(t *testing.T) {
		fromEnv, envHits := countingServer(t)
		fromAttribute, attributeHits := countingServer(t)
		t.Setenv("HOOKDECK_OUTPOST_API_BASE", fromEnv.URL)

		client := configuredClient(t, &fromAttribute.URL)
		if err := client.Outpost().Do(t.Context(), http.MethodGet, "/config", nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if *attributeHits != 1 || *envHits != 0 {
			t.Errorf("attribute server hits = %d, environment server hits = %d", *attributeHits, *envHits)
		}
	})

	t.Run("does not move the main API", func(t *testing.T) {
		outpost, outpostHits := countingServer(t)
		main, mainHits := countingServer(t)
		t.Setenv("HOOKDECK_API_BASE", main.URL)
		t.Setenv("HOOKDECK_OUTPOST_API_BASE", outpost.URL)

		client := configuredClient(t, nil)
		if err := client.Do(t.Context(), http.MethodGet, "/projects", nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if *mainHits != 1 || *outpostHits != 0 {
			t.Errorf("main API hits = %d, Outpost API hits = %d", *mainHits, *outpostHits)
		}
	})
}

func TestProviderSchema_outpostAPIBaseRejectsAPath(t *testing.T) {
	p := provider.New("test")()
	resp := &fwprovider.SchemaResponse{}
	p.Schema(t.Context(), fwprovider.SchemaRequest{}, resp)
	attribute, ok := resp.Schema.Attributes["outpost_api_base"].(schema.StringAttribute)
	if !ok || !attribute.Optional {
		t.Fatalf("outpost_api_base = %#v", resp.Schema.Attributes["outpost_api_base"])
	}
	for value, wantError := range map[string]bool{
		"https://api.outpost.example.com":      false,
		"api.outpost.example.com":              false,
		"https://api.outpost.example.com/2026": true,
	} {
		gotError := false
		for _, v := range attribute.Validators {
			r := &validator.StringResponse{}
			v.ValidateString(t.Context(), validator.StringRequest{ConfigValue: types.StringValue(value)}, r)
			gotError = gotError || r.Diagnostics.HasError()
		}
		if gotError != wantError {
			t.Errorf("%q: validation error = %v, want %v", value, gotError, wantError)
		}
	}
}
