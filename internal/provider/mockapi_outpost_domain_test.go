package provider_test

import (
	"encoding/json"
	"net/http"
)

// mockDomain is a portal custom domain in the mock Outpost API.
type mockDomain struct {
	hostname string
	active   bool
}

func (d *mockDomain) view() map[string]any {
	status, sslStatus := "pending", "pending_validation"
	if d.active {
		status, sslStatus = "active", "active"
	}
	return map[string]any{
		"id":       "dom_" + d.hostname,
		"hostname": d.hostname,
		"status":   status,
		// The CNAME target is the portal host, the same for every domain.
		"cname_target": "portal.outpost.example.net",
		"ssl": map[string]any{
			"id":        "cert_1",
			"method":    "txt",
			"status":    sslStatus,
			"txt_name":  "_acme-challenge." + d.hostname,
			"txt_value": "ssl-token",
		},
		"verification_errors": []string{},
		"ownership_verification": map[string]any{
			"type":  "txt",
			"name":  "_ownership." + d.hostname,
			"value": "ownership-token",
		},
		"created_at": mockTime,
	}
}

// handleDomain serves /config/custom_domain. GET answers 200 with null when
// there is no domain. POST on a project that has one is a 409, and so is a
// hostname registered anywhere. DELETE without a domain is a 404. Deleting a
// project does not release its hostname. The route does not check that the
// project is provisioned.
func (m *mockAPI) handleDomain(w http.ResponseWriter, r *http.Request, outpost *mockOutpost) {
	if status, faulty := outpost.domainFaults[r.Method]; faulty {
		apiError(w, status, "FAULT", "Injected", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if outpost.domain == nil {
			writeJSON(w, 200, nil)
			return
		}
		writeJSON(w, 200, outpost.domain.view())
	case http.MethodPost:
		var body struct {
			Hostname string `json:"hostname"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hostname == "" {
			apiError(w, 422, "VALIDATION_ERROR", "\"hostname\" is required", nil)
			return
		}
		if m.refuseHostnames {
			apiError(w, 422, "VALIDATION_ERROR", "\"hostname\" must be a valid hostname", nil)
			return
		}
		if outpost.domain != nil {
			apiError(w, 409, "CUSTOM_DOMAIN_ALREADY_CONFIGURED", "Custom domain already configured for the project. Delete existing domain to configure a new one", map[string]any{"hostname": outpost.domain.hostname})
			return
		}
		if m.hostnames[body.Hostname] {
			apiError(w, 409, "CUSTOM_DOMAIN_ALREADY_EXISTS", "Custom domain already exists", map[string]any{"hostname": body.Hostname})
			return
		}
		m.hostnames[body.Hostname] = true
		outpost.domain = &mockDomain{hostname: body.Hostname}
		writeJSON(w, 200, outpost.domain.view())
	case http.MethodDelete:
		if outpost.domain == nil {
			apiError(w, 404, "CUSTOM_DOMAIN_NOT_FOUND", "Custom domain not found", nil)
			return
		}
		delete(m.hostnames, outpost.domain.hostname)
		outpost.domain = nil
		writeJSON(w, 200, map[string]any{"success": true})
	default:
		apiError(w, 404, "NOT_FOUND", "Not Found", nil)
	}
}
