package project

import (
	"strings"
	"testing"

	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const domainJSON = `{"id":"d1","hostname":"portal.example.com","status":"pending","cname_target":"portal.outpost.example.net",
"ssl":{"id":"s1","method":"txt","status":"pending_validation","txt_name":"_acme-challenge.portal.example.com","txt_value":"ssl-token"},
"verification_errors":["not verified yet"],
"ownership_verification":{"type":"txt","name":"_ownership.portal.example.com","value":"ownership-token"}}`

func TestGetDomain(t *testing.T) {
	tests := []struct {
		name     string
		response response
		want     string
		wantErr  bool
	}{
		{"domain", response{200, domainJSON}, "portal.example.com", false},
		{"null body means none", response{200, "null"}, "", false},
		{"empty body means none", response{200, ""}, "", false},
		{"project without deployment means none", response{404, `{"code":"NOT_FOUND","status":404}`}, "", false},
		{"other errors are returned", response{500, "boom"}, "", true},
		{"forbidden is returned", response{403, `{"code":"INSUFFICIENT_SCOPE"}`}, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &queuedAPI{responses: []response{tc.response}}
			got, err := getDomain(t.Context(), sdkclient.Client{RawClient: api})
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.want == "" && got != nil || tc.want != "" && (got == nil || got.Hostname != tc.want) {
				t.Errorf("domain = %+v, want %q", got, tc.want)
			}
			if len(api.calls) != 1 || api.calls[0] != "GET /config/custom_domain" {
				t.Errorf("calls = %v", api.calls)
			}
		})
	}
}

func TestAddDomain(t *testing.T) {
	tests := []struct {
		name     string
		response response
		wantErr  string
	}{
		{"added", response{200, domainJSON}, ""},
		{"already configured", response{409, `{"code":"CUSTOM_DOMAIN_ALREADY_CONFIGURED","status":409,"message":"Custom domain already configured"}`}, "already has a portal custom domain"},
		{"hostname in use", response{409, `{"code":"CUSTOM_DOMAIN_ALREADY_EXISTS","status":409,"message":"Custom domain already exists"}`}, "portal.example.com is already in use by another project"},
		{"invalid hostname is the API error", response{422, `{"code":"VALIDATION_ERROR","status":422,"message":"invalid hostname"}`}, "invalid hostname"},
		{"other error keeps the API message", response{500, `{"code":"INTERNAL","message":"boom"}`}, "boom"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &queuedAPI{responses: []response{tc.response}}
			got, err := addDomain(t.Context(), sdkclient.Client{RawClient: api}, "portal.example.com")
			switch {
			case tc.wantErr == "" && (err != nil || got == nil):
				t.Fatalf("got %+v, %v", got, err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("err = %v, want it to contain %q", err, tc.wantErr)
			}
			if len(api.calls) != 1 || api.calls[0] != "POST /config/custom_domain" || api.bodies[0] != `{"hostname":"portal.example.com"}` {
				t.Errorf("calls = %v, bodies = %v", api.calls, api.bodies)
			}
		})
	}
}

func TestDeleteDomain(t *testing.T) {
	tests := []struct {
		name     string
		response response
		wantErr  bool
	}{
		{"deleted", response{200, `{"success":true}`}, false},
		{"no domain", response{404, `{"code":"CUSTOM_DOMAIN_NOT_FOUND","status":404}`}, false},
		{"no deployment", response{404, `{"code":"NOT_FOUND","status":404}`}, false},
		{"project gone", response{404, `{"code":"NOT_FOUND","status":404,"message":"Team Not Found"}`}, false},
		{"server error", response{500, "boom"}, true},
		{"forbidden", response{403, `{"code":"INSUFFICIENT_SCOPE"}`}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			api := &queuedAPI{responses: []response{tc.response}}
			client := outpostAPI(sdkclient.Client{OutpostRawClient: api}, "tm_1")
			if err := deleteDomain(t.Context(), client); (err != nil) != tc.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if len(api.calls) != 1 || api.calls[0] != "DELETE /config/custom_domain" {
				t.Errorf("calls = %v", api.calls)
			}
		})
	}
}

func asList(t *testing.T, v attr.Value) types.List {
	t.Helper()
	list, ok := v.(types.List)
	if !ok {
		t.Fatalf("%v is not a list", v)
	}
	return list
}

func asObject(t *testing.T, v attr.Value) types.Object {
	t.Helper()
	obj, ok := v.(types.Object)
	if !ok {
		t.Fatalf("%v is not an object", v)
	}
	return obj
}

func TestStatusObject(t *testing.T) {
	t.Run("no domain is null", func(t *testing.T) {
		var d *customDomain
		got, diags := d.statusObject(t.Context())
		if diags.HasError() || !got.IsNull() {
			t.Errorf("got %v, %v", got, diags)
		}
	})

	t.Run("mirrors the API response", func(t *testing.T) {
		d, err := addDomain(t.Context(), sdkclient.Client{RawClient: &queuedAPI{responses: []response{{200, domainJSON}}}}, "portal.example.com")
		if err != nil {
			t.Fatal(err)
		}
		got, diags := d.statusObject(t.Context())
		if diags.HasError() {
			t.Fatal(diags)
		}
		attrs := got.Attributes()
		for key, want := range map[string]string{"status": "pending", "cname_target": "portal.outpost.example.net"} {
			if attrs[key] != types.StringValue(want) {
				t.Errorf("%s = %v, want %s", key, attrs[key], want)
			}
		}
		for key, want := range map[string]map[string]string{
			"ssl": {
				"status":    "pending_validation",
				"method":    "txt",
				"txt_name":  "_acme-challenge.portal.example.com",
				"txt_value": "ssl-token",
			},
			"ownership_verification": {
				"type":  "txt",
				"name":  "_ownership.portal.example.com",
				"value": "ownership-token",
			},
		} {
			fields := asObject(t, attrs[key]).Attributes()
			for field, value := range want {
				if fields[field] != types.StringValue(value) {
					t.Errorf("%s.%s = %v, want %s", key, field, fields[field], value)
				}
			}
		}
		errs := asList(t, attrs["verification_errors"]).Elements()
		if len(errs) != 1 || errs[0] != types.StringValue("not verified yet") {
			t.Errorf("verification_errors = %v", errs)
		}
	})

	t.Run("values are kept as the API returns them", func(t *testing.T) {
		d := &customDomain{}
		d.OwnershipVerification = &struct {
			Type  string `json:"type"`
			Name  string `json:"name"`
			Value string `json:"value"`
		}{Type: "TXT", Name: "_Owner.Portal.example.com", Value: "Token"}
		got, diags := d.statusObject(t.Context())
		if diags.HasError() {
			t.Fatal(diags)
		}
		fields := asObject(t, got.Attributes()["ownership_verification"]).Attributes()
		if fields["type"] != types.StringValue("TXT") || fields["name"] != types.StringValue("_Owner.Portal.example.com") || fields["value"] != types.StringValue("Token") {
			t.Errorf("ownership_verification = %v", fields)
		}
	})

	t.Run("no ownership record is null, the error list is never null", func(t *testing.T) {
		d := &customDomain{Hostname: "portal.example.com", Status: "active"}
		got, diags := d.statusObject(t.Context())
		if diags.HasError() {
			t.Fatal(diags)
		}
		attrs := got.Attributes()
		if !asObject(t, attrs["ownership_verification"]).IsNull() {
			t.Errorf("ownership_verification = %v, want null", attrs["ownership_verification"])
		}
		if list := asList(t, attrs["verification_errors"]); list.IsNull() || len(list.Elements()) != 0 {
			t.Errorf("verification_errors = %v, want an empty list", list)
		}
	})

	t.Run("attribute types match the schema", func(t *testing.T) {
		d := &customDomain{}
		got, _ := d.statusObject(t.Context())
		want := domainStatusTypes
		for k, typ := range got.AttributeTypes(t.Context()) {
			if !typ.Equal(want[k]) {
				t.Errorf("%s: type %v, want %v", k, typ, want[k])
			}
		}
	})
}

func TestDomainValues(t *testing.T) {
	tests := []struct {
		name string
		live *customDomain
		want types.String
	}{
		{"no domain", nil, types.StringNull()},
		{"live domain", &customDomain{Hostname: "portal.example.com"}, types.StringValue("portal.example.com")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, status, diags := domainValues(t.Context(), tc.live)
			if diags.HasError() || !got.Equal(tc.want) || status.IsNull() != (tc.live == nil) {
				t.Errorf("hostname = %v (want %v), status null = %v, diags = %v", got, tc.want, status.IsNull(), diags)
			}
		})
	}
}
