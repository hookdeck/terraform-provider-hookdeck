// Package projectscope resolves which Hookdeck project a request targets.
//
// An organization API key (prefix hd_org_) can address several projects, so
// every project-scoped resource needs a project id: from the resource, then
// the provider default. A project API key addresses exactly one project and
// any explicit project id must match it.
package projectscope

import "errors"

// KeyKind is the kind of API key the provider was configured with.
type KeyKind int

const (
	// KeyKindProject is a project API key (prefix hd_, or no prefix).
	KeyKindProject KeyKind = iota
	// KeyKindOrganization is an organization API key (prefix hd_org_).
	KeyKindOrganization
)

var (
	// ErrProjectRequired is returned when an organization key has no
	// project to target.
	ErrProjectRequired = errors.New("project_id is required")
	// ErrProjectMismatch is returned when a project key is used with a
	// different project id.
	ErrProjectMismatch = errors.New("project_id does not match the API key's project")
	// ErrInvalidImportID is returned for import ids that are neither
	// <id> nor <project_id>/<id>.
	ErrInvalidImportID = errors.New("import id must be <id> or <project_id>/<id>")
)

// KindOfKey derives the key kind from the key's prefix.
func KindOfKey(apiKey string) KeyKind {
	panic("not implemented")
}

// Resolve returns the project id requests should target, or "" when the
// key's own project applies and no header is needed.
//
// keyProject is called at most once, only when the resolution needs the
// project key's own project id to validate an explicit project_id.
func Resolve(kind KeyKind, resourceProjectID, providerProjectID string, keyProject func() (string, error)) (string, error) {
	panic("not implemented")
}

// ParseImportID splits "<project_id>/<id>" or "<id>".
func ParseImportID(importID string) (projectID, id string, err error) {
	panic("not implemented")
}
