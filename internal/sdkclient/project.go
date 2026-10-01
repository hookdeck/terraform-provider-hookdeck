package sdkclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"terraform-provider-hookdeck/internal/projectscope"
)

// ProjectHeader selects the project a request acts on. The API calls a
// project a team.
const ProjectHeader = "x-hookdeck-team-id"

// ProjectAccessError is the API refusing the project a request was scoped
// to, as opposed to the resource the request named.
type ProjectAccessError struct {
	ProjectID string
	KeyKind   projectscope.KeyKind
	Status    int
	Body      string
	// Hint is appended to the message; it tells the caller's user what to
	// do next.
	Hint string
}

func (e *ProjectAccessError) Error() string {
	var reason string
	switch {
	case e.Status == http.StatusNotFound:
		reason = "The project does not exist or belongs to another organization."
	case e.Status == http.StatusForbidden:
		reason = "The organization API key has no grant on the project, or the grant lacks the scope this request needs."
	case e.KeyKind == projectscope.KeyKindProject:
		reason = "A project API key only reaches its own project; the key belongs to another project or is invalid."
	default:
		reason = "The project belongs to another organization, or the API key is invalid."
	}
	msg := fmt.Sprintf("The API key cannot access project %s. %s", e.ProjectID, reason)
	if e.Hint != "" {
		msg += " " + e.Hint
	}
	if body := strings.TrimSpace(e.Body); body != "" {
		msg += fmt.Sprintf("\n\nAPI response (%d): %s", e.Status, body)
	}
	return msg
}

// WithProject returns a client whose requests act on projectID. A response
// that refuses the project is returned as *ProjectAccessError carrying hint.
func (c Client) WithProject(projectID, hint string) Client {
	scoped := c
	scoped.RawClient = projectHeaderClient{
		inner:     c.RawClient,
		projectID: projectID,
		keyKind:   c.Scope.KeyKind,
		hint:      hint,
	}
	return scoped
}

type projectHeaderClient struct {
	inner     RawClientInterface
	projectID string
	keyKind   projectscope.KeyKind
	hint      string
}

func (p projectHeaderClient) SendRequest(ctx context.Context, method, path string, opts *RequestOptions) (*http.Response, error) {
	var scoped RequestOptions
	if opts != nil {
		scoped = *opts
	}
	headers := http.Header{}
	for k, v := range scoped.Headers {
		headers[k] = v
	}
	headers.Set(ProjectHeader, p.projectID)
	scoped.Headers = headers

	resp, err := p.inner.SendRequest(ctx, method, path, &scoped)
	if err != nil {
		return nil, err
	}
	if err := p.projectRefused(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// projectRefused reports a response that refuses the project itself: 401
// (a project key on another project, or a project in another organization),
// 404 for the project (an organization key naming a project that does not
// exist) and 403 for a missing grant. Any other response is left readable.
func (p projectHeaderClient) projectRefused(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
	default:
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	var parsed struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &parsed)

	refused := false
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		refused = true
	case http.StatusForbidden:
		refused = p.keyKind == projectscope.KeyKindOrganization && parsed.Code == "INSUFFICIENT_SCOPE"
	case http.StatusNotFound:
		refused = parsed.Message == "Team Not Found" || parsed.Data.ID == p.projectID
	}
	if !refused {
		return nil
	}
	return &ProjectAccessError{
		ProjectID: p.projectID,
		KeyKind:   p.keyKind,
		Status:    resp.StatusCode,
		Body:      string(body),
		Hint:      p.hint,
	}
}

// KeyProjectID returns the project a project API key belongs to.
func (c Client) KeyProjectID(ctx context.Context) (string, error) {
	var projects []struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodGet, "/projects", nil, nil, &projects); err != nil {
		return "", err
	}
	if len(projects) != 1 || projects[0].ID == "" {
		return "", fmt.Errorf("expected the API key's one project, got %d projects", len(projects))
	}
	return projects[0].ID, nil
}

type projectChecks struct {
	mu      sync.Mutex
	results map[string]error
}

// CheckProject reports whether an organization API key can see projectID.
// It returns *ProjectAccessError when the project is not visible and nil
// when it is, or when the key may not read projects and the answer is
// unknown. Results are kept for the life of the provider instance.
func (c Client) CheckProject(ctx context.Context, projectID, hint string) error {
	if c.projectChecks != nil {
		c.projectChecks.mu.Lock()
		defer c.projectChecks.mu.Unlock()
		if err, ok := c.projectChecks.results[projectID]; ok {
			return err
		}
	}
	err := c.Do(ctx, http.MethodGet, "/projects/"+projectID, nil, nil, nil)
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Status {
		case http.StatusForbidden:
			err = nil
		case http.StatusNotFound:
			err = &ProjectAccessError{
				ProjectID: projectID,
				KeyKind:   c.Scope.KeyKind,
				Status:    apiErr.Status,
				Body:      apiErr.Body,
				Hint:      hint,
			}
		}
	}
	var accessErr *ProjectAccessError
	if c.projectChecks != nil && (err == nil || errors.As(err, &accessErr)) {
		c.projectChecks.results[projectID] = err
	}
	return err
}
