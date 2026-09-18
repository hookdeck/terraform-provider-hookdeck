package sdkclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"terraform-provider-hookdeck/internal/projectscope"
)

// APIVersion is the Hookdeck API version every request targets.
const APIVersion = "2026-09-01"

// ProjectHeader selects the project an organization API key acts on.
const ProjectHeader = "x-hookdeck-team-id"

// keyProjectLookup resolves, once, the project a project API key belongs to.
type keyProjectLookup struct {
	once sync.Once
	id   string
	err  error
}

// KeyProjectID returns the project the configured project API key belongs
// to. The lookup runs once per provider instance.
func (c Client) KeyProjectID(ctx context.Context) (string, error) {
	if c.keyProject == nil {
		return "", fmt.Errorf("client was not created with InitHookdeckSDKClient")
	}
	c.keyProject.once.Do(func() {
		c.keyProject.id, c.keyProject.err = lookupKeyProject(ctx, c.RawClient)
	})
	return c.keyProject.id, c.keyProject.err
}

func lookupKeyProject(ctx context.Context, raw RawClientInterface) (string, error) {
	resp, err := raw.SendRequest(ctx, "GET", "/"+APIVersion+"/projects", nil)
	if err != nil {
		return "", fmt.Errorf("looking up the API key's project: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("looking up the API key's project: %w", err)
	}
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("looking up the API key's project: %d %s", resp.StatusCode, string(body))
	}
	var projects []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &projects); err != nil {
		return "", fmt.Errorf("looking up the API key's project: %w", err)
	}
	if len(projects) != 1 || projects[0].ID == "" {
		return "", fmt.Errorf("looking up the API key's project: expected one project, got %d", len(projects))
	}
	return projects[0].ID, nil
}

// ForProject returns a client whose requests target the resolved project.
// resourceProjectID is the resource's own project_id, "" when unset.
func (c Client) ForProject(ctx context.Context, resourceProjectID string) (Client, error) {
	projectID, err := projectscope.Resolve(c.KeyKind, resourceProjectID, c.DefaultProjectID, func() (string, error) {
		return c.KeyProjectID(ctx)
	})
	if err != nil {
		return Client{}, err
	}
	if projectID == "" {
		return c, nil
	}
	return c.WithProject(projectID), nil
}

// WithProject returns a client that sends every request with the project
// header set to projectID. No validation is performed.
func (c Client) WithProject(projectID string) Client {
	scoped := c
	scoped.RawClient = projectHeaderClient{inner: c.RawClient, projectID: projectID}
	return scoped
}

type projectHeaderClient struct {
	inner     RawClientInterface
	projectID string
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
	return p.inner.SendRequest(ctx, method, path, &scoped)
}
