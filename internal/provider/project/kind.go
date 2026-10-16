package project

import "fmt"

// kind is one project type of the Hookdeck API, exposed as its own
// resource and data source.
type kind struct {
	// typeSuffix is the type name without the provider prefix.
	typeSuffix string
	// apiType is the project's type on the API.
	apiType string
	label   string
}

var gateway = kind{typeSuffix: "_gateway_project", apiType: "event_gateway", label: "Event Gateway"}

func (k kind) typeName(providerTypeName string) string {
	return providerTypeName + k.typeSuffix
}

func (k kind) wrongType(id, apiType string) string {
	return fmt.Sprintf("Project %s has type %q. %s manages %s projects (type %q) only.", id, apiType, k.typeName("hookdeck"), k.label, k.apiType)
}
