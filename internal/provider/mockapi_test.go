package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"terraform-provider-hookdeck/internal/sdkclient"
)

// mockAPI is an in-process Hookdeck API with the project and source routes
// and the key and project rules of the real one. It covers the cases the
// test credentials cannot reach: keys without projects.read, lost grants,
// other project types.
type mockAPI struct {
	mu       sync.Mutex
	server   *httptest.Server
	keys     map[string]*mockKey
	projects map[string]*mockProject
	sources  map[string]map[string]any
	// destinations are stored like sources: the object as the API returns it.
	destinations map[string]map[string]any
	// strippedKeys are top-level config keys the API drops on write, the way
	// it drops keys it does not know for the type.
	strippedKeys map[string]bool
	seq          int
	requests     []mockRequest
}

type mockKey struct {
	org string
	// project is set for a project key and empty for an organization key.
	project string
	// noProjectsRead removes the projects.read scope.
	noProjectsRead bool
	// grants limits an organization key to these projects; nil is all.
	grants map[string]bool
}

type mockProject struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	OrganizationID      string   `json:"organization_id"`
	HeadersPrefix       *string  `json:"headers_prefix"`
	NotificationMethods []string `json:"notification_methods"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type mockRequest struct {
	key, method, path, project string
	body                       map[string]any
}

const (
	mockOrg  = "org_1"
	mockTime = "2026-09-01T00:00:00.000Z"
)

var apiVersionPrefix = regexp.MustCompile(`^/\d{4}-\d{2}-\d{2}`)

// newMockAPI starts the mock and points the provider under test at it.
func newMockAPI(t *testing.T) *mockAPI {
	t.Helper()
	m := &mockAPI{
		keys:     map[string]*mockKey{},
		projects: map[string]*mockProject{},
		sources:  map[string]map[string]any{},

		destinations: map[string]map[string]any{},
		strippedKeys: map[string]bool{},
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.handle))
	t.Cleanup(m.server.Close)
	t.Setenv("HOOKDECK_API_BASE", m.server.URL)
	return m
}

func (m *mockAPI) addProject(id, name, projectType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[id] = &mockProject{ID: id, Name: name, Type: projectType, OrganizationID: mockOrg, NotificationMethods: []string{}, CreatedAt: mockTime, UpdatedAt: mockTime}
}

func (m *mockAPI) addKey(apiKey string, key mockKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[apiKey] = &key
}

func (m *mockAPI) update(change func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	change()
}

// requestsTo returns the recorded requests whose path starts with prefix.
func (m *mockAPI) requestsTo(prefix string) []mockRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []mockRequest
	for _, r := range m.requests {
		if strings.HasPrefix(r.path, prefix) {
			out = append(out, r)
		}
	}
	return out
}

// copyJSON returns a deep copy of a decoded JSON object, so a recorded request
// does not change when the mock later merges the same body into stored data.
func copyJSON(v map[string]any) map[string]any {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return v
	}
	return out
}

func (m *mockAPI) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Any API version: released providers use older ones.
	path := apiVersionPrefix.ReplaceAllString(r.URL.Path, "")
	apiKey := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	header := r.Header.Get(sdkclient.ProjectHeader)
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	m.requests = append(m.requests, mockRequest{key: apiKey, method: r.Method, path: path, project: header, body: copyJSON(body)})

	key, ok := m.keys[apiKey]
	if !ok {
		unauthorized(w)
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	id := ""
	if len(parts) > 1 {
		id = parts[1]
	}
	switch parts[0] {
	case "projects":
		m.handleProjects(w, r.Method, key, id, body)
	case "sources":
		m.handleObjects(w, r.Method, key, header, id, body, sourceKind)
	case "destinations":
		m.handleObjects(w, r.Method, key, header, id, body, destinationKind)
	default:
		apiError(w, 404, "NOT_FOUND", "Not Found", nil)
	}
}

func (m *mockAPI) handleProjects(w http.ResponseWriter, method string, key *mockKey, id string, body map[string]any) {
	organizationKey := key.project == ""
	visible := func(p *mockProject) bool {
		if !organizationKey {
			return p.ID == key.project
		}
		return p.OrganizationID == key.org && (key.grants == nil || key.grants[p.ID])
	}

	if id == "" {
		switch method {
		case http.MethodGet:
			if key.noProjectsRead {
				insufficientScope(w, "projects.read")
				return
			}
			list := []*mockProject{}
			for _, p := range m.projects {
				if visible(p) {
					list = append(list, p)
				}
			}
			writeJSON(w, 200, list)
		case http.MethodPost:
			if !organizationKey {
				unauthorized(w)
				return
			}
			m.seq++
			name, _ := body["name"].(string)
			projectType, _ := body["type"].(string)
			p := &mockProject{ID: fmt.Sprintf("tm_new%d", m.seq), Name: name, Type: projectType, OrganizationID: key.org, NotificationMethods: []string{}, CreatedAt: mockTime, UpdatedAt: mockTime}
			m.projects[p.ID] = p
			if key.grants != nil {
				key.grants[p.ID] = true
			}
			writeJSON(w, 200, p)
		}
		return
	}

	p, ok := m.projects[id]
	if method != http.MethodGet && !organizationKey {
		unauthorized(w)
		return
	}
	// The API checks the scope for the project named in the path before it
	// looks the project up: a key limited to other projects gets 403 for a
	// project that does not exist too.
	if key.noProjectsRead || (organizationKey && key.grants != nil && !key.grants[id]) {
		insufficientScope(w, "projects.read")
		return
	}
	if !ok || p.OrganizationID != key.org || (!organizationKey && p.ID != key.project) {
		apiError(w, 404, "NOT_FOUND", "Not Found", map[string]any{"id": id})
		return
	}
	switch method {
	case http.MethodGet:
		writeJSON(w, 200, p)
	case http.MethodPut:
		if name, ok := body["name"].(string); ok {
			p.Name = name
		}
		if prefix, present := body["headers_prefix"]; present {
			p.HeadersPrefix = nil
			if s, ok := prefix.(string); ok {
				p.HeadersPrefix = &s
			}
		}
		if methods, ok := body["notification_methods"].([]any); ok {
			p.NotificationMethods = []string{}
			for _, v := range methods {
				p.NotificationMethods = append(p.NotificationMethods, fmt.Sprint(v))
			}
		}
		writeJSON(w, 200, p)
	case http.MethodDelete:
		delete(m.projects, id)
		for _, store := range []map[string]map[string]any{m.sources, m.destinations} {
			for objectID, object := range store {
				if object["team_id"] == id {
					delete(store, objectID)
				}
			}
		}
		writeJSON(w, 200, map[string]any{"id": id})
	}
}

// project resolves the project a request acts on, the way the API does.
func (m *mockAPI) project(w http.ResponseWriter, key *mockKey, header string) (string, bool) {
	if key.project != "" {
		if header != "" && header != key.project {
			unauthorized(w)
			return "", false
		}
		return key.project, true
	}
	if header == "" {
		unauthorized(w)
		return "", false
	}
	p, ok := m.projects[header]
	if !ok {
		apiError(w, 404, "NOT_FOUND", "Team Not Found", map[string]any{"id": header})
		return "", false
	}
	if p.OrganizationID != key.org {
		unauthorized(w)
		return "", false
	}
	if key.grants != nil && !key.grants[header] {
		insufficientScope(w, "gateway.sources.read")
		return "", false
	}
	return header, true
}

// mockKind describes one kind of project object the mock stores.
type mockKind struct {
	label, idPrefix string
	store           func(m *mockAPI) map[string]map[string]any
	// defaults fills in the config keys the API adds when they are missing.
	defaults func(objectType any, config map[string]any)
	// merge applies one key of a config update; false falls back to the
	// default behaviour: null removes the key, a value replaces it.
	merge func(stored map[string]any, key string, value any) bool
	// extra adds the fields only this kind has.
	extra func(object map[string]any, seq int)
}

var sourceKind = mockKind{
	label:    "Source",
	idPrefix: "src_",
	store:    func(m *mockAPI) map[string]map[string]any { return m.sources },
	defaults: func(objectType any, config map[string]any) {
		setDefault(config, "auth_type", nil)
		setDefault(config, "auth", nil)
		setDefault(config, "custom_response", nil)
		if objectType == "HTTP" || objectType == "WEBHOOK" {
			setDefault(config, "allowed_http_methods", []any{"POST", "PUT", "PATCH", "DELETE"})
		}
	},
	extra: func(object map[string]any, seq int) {
		object["url"] = fmt.Sprintf("https://hkdk.events/%d", seq)
	},
}

var destinationKind = mockKind{
	label:    "Destination",
	idPrefix: "des_",
	store:    func(m *mockAPI) map[string]map[string]any { return m.destinations },
	defaults: func(objectType any, config map[string]any) {
		setDefault(config, "auth_type", "HOOKDECK_SIGNATURE")
		setDefault(config, "auth", map[string]any{})
		setDefault(config, "http_method", nil)
		setDefault(config, "path_forwarding_disabled", false)
		setDefault(config, "delivery_policy", map[string]any{"rate": nil, "period": nil, "groups": nil})
	},
	// Delivery policy members are merged one by one.
	merge: func(stored map[string]any, key string, value any) bool {
		if key != "delivery_policy" {
			return false
		}
		members, ok := value.(map[string]any)
		policy, _ := stored[key].(map[string]any)
		if !ok || policy == nil {
			return false
		}
		for member, v := range members {
			policy[member] = v
		}
		return true
	},
}

func setDefault(config map[string]any, key string, value any) {
	if _, ok := config[key]; !ok {
		config[key] = value
	}
}

// writeConfig applies a create or update body's config to the stored object.
func (m *mockAPI) writeConfig(kind mockKind, object map[string]any, config map[string]any) {
	stored, _ := object["config"].(map[string]any)
	if stored == nil {
		stored = map[string]any{}
		object["config"] = stored
	}
	for k, v := range config {
		if kind.merge != nil && kind.merge(stored, k, v) {
			continue
		}
		if v == nil {
			delete(stored, k)
		} else {
			stored[k] = v
		}
	}
	for k := range m.strippedKeys {
		delete(stored, k)
	}
	kind.defaults(object["type"], stored)
}

// create stores a new object the way a create request does.
func (m *mockAPI) create(kind mockKind, project string, body map[string]any) map[string]any {
	m.seq++
	object := map[string]any{
		"id":          fmt.Sprintf("%s%d", kind.idPrefix, m.seq),
		"name":        body["name"],
		"type":        body["type"],
		"description": body["description"],
		"team_id":     project,
		"config":      map[string]any{},
		"disabled_at": nil,
		"created_at":  mockTime,
		"updated_at":  mockTime,
	}
	if kind.extra != nil {
		kind.extra(object, m.seq)
	}
	config, _ := body["config"].(map[string]any)
	m.writeConfig(kind, object, config)
	kind.store(m)[object["id"].(string)] = object //nolint:forcetypeassert
	return object
}

func (m *mockAPI) handleObjects(w http.ResponseWriter, method string, key *mockKey, header, id string, body map[string]any, kind mockKind) {
	project, ok := m.project(w, key, header)
	if !ok {
		return
	}
	store := kind.store(m)
	if id == "" && method == http.MethodPost {
		writeJSON(w, 200, m.create(kind, project, body))
		return
	}
	if id == "" && method == http.MethodGet {
		models := []map[string]any{}
		for _, object := range store {
			if object["team_id"] == project {
				models = append(models, object)
			}
		}
		writeJSON(w, 200, map[string]any{"models": models, "count": len(models)})
		return
	}
	object, ok := store[id]
	if !ok || object["team_id"] != project {
		apiError(w, 404, "NOT_FOUND", kind.label+" Not Found", map[string]any{"id": id})
		return
	}
	switch method {
	case http.MethodPut:
		for _, field := range []string{"name", "type", "description"} {
			if v, present := body[field]; present {
				object[field] = v
			}
		}
		if config, ok := body["config"].(map[string]any); ok {
			m.writeConfig(kind, object, config)
		}
	case http.MethodDelete:
		delete(store, id)
	}
	writeJSON(w, 200, object)
}

// setSourceConfig changes one config key of a source the way an edit in the
// dashboard does: no request reaches the mock's recorded requests.
func (m *mockAPI) setSourceConfig(id, key string, value any) {
	m.update(func() { m.sources[id]["config"].(map[string]any)[key] = value }) //nolint:forcetypeassert
}

// setDestinationConfig is setSourceConfig for a destination.
func (m *mockAPI) setDestinationConfig(id, key string, value any) {
	m.update(func() { m.destinations[id]["config"].(map[string]any)[key] = value }) //nolint:forcetypeassert
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiError(w http.ResponseWriter, status int, code, message string, data any) {
	writeJSON(w, status, map[string]any{"code": code, "status": status, "message": message, "data": data})
}

func insufficientScope(w http.ResponseWriter, required string) {
	apiError(w, 403, "INSUFFICIENT_SCOPE", "The API key does not have the required scope for this operation", map[string]any{"required": required})
}

func unauthorized(w http.ResponseWriter) {
	w.WriteHeader(401)
	_, _ = w.Write([]byte("Unauthorized"))
}

// createSource adds a source without a request, like one made outside
// Terraform, and returns its id.
func (m *mockAPI) createSource(project, name, sourceType string, config map[string]any) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.create(sourceKind, project, map[string]any{"name": name, "type": sourceType, "config": config})["id"].(string) //nolint:forcetypeassert
}

// createDestination is createSource for a destination.
func (m *mockAPI) createDestination(project, name, destinationType string, config map[string]any) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.create(destinationKind, project, map[string]any{"name": name, "type": destinationType, "config": config})["id"].(string) //nolint:forcetypeassert
}
