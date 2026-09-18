// Package shared holds the pieces every project-scoped resource and data
// source uses: the project_id attribute, client resolution and import id
// parsing.
package shared

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/sdkclient"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const projectIDDescription = "ID of the project the resource belongs to. Defaults to the provider `project_id`, then to the API key's own project. Required with an organization API key when the provider sets no default. Changing it replaces the resource."

// ProjectIDResourceAttribute is the project_id attribute for resources.
func ProjectIDResourceAttribute() resourceschema.StringAttribute {
	return resourceschema.StringAttribute{
		Optional:    true,
		Description: projectIDDescription,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// ProjectIDDataSourceAttribute is the project_id attribute for data sources.
func ProjectIDDataSourceAttribute() datasourceschema.StringAttribute {
	return datasourceschema.StringAttribute{
		Optional:    true,
		Description: "ID of the project to read from. Defaults to the provider `project_id`, then to the API key's own project.",
	}
}

// ClientFromProviderData asserts the provider-configured client.
func ClientFromProviderData(providerData any) (sdkclient.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	client, ok := providerData.(sdkclient.Client)
	if !ok {
		diags.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected sdkclient.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
	}
	return client, diags
}

// ClientFor resolves the project for a resource and returns a client scoped
// to it.
func ClientFor(ctx context.Context, client sdkclient.Client, projectID types.String) (sdkclient.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	scoped, err := client.ForProject(ctx, projectID.ValueString())
	if err != nil {
		switch {
		case errors.Is(err, projectscope.ErrProjectRequired):
			diags.AddAttributeError(path.Root("project_id"), "Missing project",
				"The provider is configured with an organization API key. Set project_id on the resource or on the provider.")
		case errors.Is(err, projectscope.ErrProjectMismatch):
			diags.AddAttributeError(path.Root("project_id"), "Project mismatch",
				"project_id does not match the API key's project. Remove project_id, or use an organization API key.")
		default:
			diags.AddError("Error resolving project", err.Error())
		}
	}
	return scoped, diags
}

// ImportState accepts "<id>" or "<project_id>/<id>" and writes the id to
// idAttribute and, when present, the project to project_id.
func ImportState(ctx context.Context, idAttribute string, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, id, err := projectscope.ParseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import id", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(idAttribute), id)...)
	if projectID != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), projectID)...)
	}
}
