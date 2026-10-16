// Package projectscope holds the rules for which Hookdeck project a resource
// belongs to.
//
// A provider configuration works in one of two modes. In single-project mode
// (a project API key, or a provider project_id) every resource is in that one
// project. In explicit mode (an organization API key without a provider
// project_id) every resource names its project.
package projectscope

import (
	"errors"
	"fmt"
	"strings"
)

// KeyKind is the kind of API key the provider was configured with.
type KeyKind int

const (
	// KeyKindProject is a project API key (prefix hd_, or no prefix).
	KeyKindProject KeyKind = iota
	// KeyKindOrganization is an organization API key (prefix hd_org_).
	KeyKindOrganization
)

const OrganizationKeyPrefix = "hd_org_"

func KindOfKey(apiKey string) KeyKind {
	if strings.HasPrefix(apiKey, OrganizationKeyPrefix) {
		return KeyKindOrganization
	}
	return KeyKindProject
}

var (
	// ErrProjectRequired is returned in explicit mode when a resource does
	// not name its project.
	ErrProjectRequired = errors.New("project_id is required")
	// ErrInvalidImportID is returned for import IDs that are neither
	// <id> nor <project_id>/<id>.
	ErrInvalidImportID = errors.New("import ID must be <id> or <project_id>/<id>")
)

// MismatchError is returned in single-project mode when a resource names a
// project other than the provider's.
type MismatchError struct {
	Provider   string
	Configured string
}

func (e *MismatchError) Error() string {
	return fmt.Sprintf("project_id %s does not match the provider's project %s", e.Configured, e.Provider)
}

// UnreachableError is returned when a resource is recorded in a project the
// project API key cannot act on.
type UnreachableError struct {
	Stored string
	Target string
}

func (e *UnreachableError) Error() string {
	return fmt.Sprintf("the resource is in project %s, the provider targets project %s", e.Stored, e.Target)
}

// Action is what a plan does about a resource's project.
type Action int

const (
	// ActionKeep leaves the resource in its project.
	ActionKeep Action = iota
	// ActionReplace deletes the resource in its project and creates it in
	// the target project.
	ActionReplace
)

// Scope is what the provider configuration says about projects.
type Scope struct {
	KeyKind KeyKind
	// ProjectID is the provider project_id, else the project API key's own
	// project when it is known. Empty in explicit mode.
	ProjectID string
}

func (s Scope) SingleProject() bool {
	return s.KeyKind == KeyKindProject || s.ProjectID != ""
}

// Target returns the project a resource or data source configured with
// project_id configured ("" when unset) belongs to. It returns "" for a
// project API key whose project is not known yet.
func (s Scope) Target(configured string) (string, error) {
	if !s.SingleProject() {
		if configured == "" {
			return "", ErrProjectRequired
		}
		return configured, nil
	}
	if s.ProjectID == "" {
		return configured, nil
	}
	if configured != "" && configured != s.ProjectID {
		return "", &MismatchError{Provider: s.ProjectID, Configured: configured}
	}
	return s.ProjectID, nil
}

// Reach reports whether the key can act on the project a resource is
// recorded in.
func (s Scope) Reach(stored string) error {
	if s.KeyKind == KeyKindProject && s.ProjectID != "" && stored != s.ProjectID {
		return &UnreachableError{Stored: stored, Target: s.ProjectID}
	}
	return nil
}

// Plan returns the project a resource ends up in and what gets it there.
// stored is the project recorded in state, "" for a new resource or for
// state that does not record one.
func (s Scope) Plan(configured, stored string) (string, Action, error) {
	target, err := s.Target(configured)
	if err != nil {
		return "", ActionKeep, err
	}
	if stored == "" {
		return target, ActionKeep, nil
	}
	if target == "" || target == stored {
		return stored, ActionKeep, nil
	}
	if s.KeyKind == KeyKindProject {
		return "", ActionKeep, &UnreachableError{Stored: stored, Target: target}
	}
	return target, ActionReplace, nil
}

// ParseImportID splits "<project_id>/<id>" or "<id>".
func ParseImportID(importID string) (projectID, id string, err error) {
	parts := strings.Split(importID, "/")
	switch len(parts) {
	case 1:
		if parts[0] == "" {
			return "", "", ErrInvalidImportID
		}
		return "", parts[0], nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return "", "", ErrInvalidImportID
		}
		return parts[0], parts[1], nil
	default:
		return "", "", ErrInvalidImportID
	}
}
