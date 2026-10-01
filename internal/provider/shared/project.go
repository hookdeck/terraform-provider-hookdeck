// Package shared holds helpers common to the provider's resources and data
// sources.
package shared

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	projectIDAttribute = "project_id"
	teamIDAttribute    = "team_id"

	storedProjectHint = "The resource is recorded in that project and nothing was changed. To stop managing it, remove it from state (terraform state rm, or a removed block). To move it to another project, use an organization API key with access to both projects."
	targetProjectHint = "Check project_id and the API key's access to that project."
)

func ProjectIDResourceAttribute() resourceschema.StringAttribute {
	return resourceschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "ID of the project the resource belongs to. With a project API key or a provider `project_id`, every resource is in that project and this can be omitted. With an organization API key and no provider `project_id`, it is required. Changing it replaces the resource.",
		Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
	}
}

func ProjectIDDataSourceAttribute() datasourceschema.StringAttribute {
	return datasourceschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "ID of the project to read from. With a project API key or a provider `project_id`, that project is used and this can be omitted. With an organization API key and no provider `project_id`, it is required.",
		Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
	}
}

// StoredProject returns the project a resource's state records: project_id,
// or team_id for state written before project_id existed.
func StoredProject(projectID, teamID types.String) string {
	if v := projectID.ValueString(); v != "" {
		return v
	}
	return teamID.ValueString()
}

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

// ProjectDiagnostics is the one mapping from a project resolution error to
// what the user sees.
func ProjectDiagnostics(err error) diag.Diagnostics {
	var diags diag.Diagnostics
	var mismatch *projectscope.MismatchError
	var unreachable *projectscope.UnreachableError
	var access *sdkclient.ProjectAccessError
	switch {
	case errors.Is(err, projectscope.ErrProjectRequired):
		diags.AddAttributeError(path.Root(projectIDAttribute), "Missing project_id",
			"project_id is required: the provider is configured with an organization API key and no project_id, so every resource and data source names its project. To manage a single project instead, set project_id on the provider.")
	case errors.As(err, &mismatch):
		diags.AddAttributeError(path.Root(projectIDAttribute), "Project mismatch",
			fmt.Sprintf("project_id is %s, but this provider configuration manages the single project %s. Remove project_id here or set it to %s. To manage several projects with one provider configuration, use an organization API key without a provider project_id.",
				mismatch.Configured, mismatch.Provider, mismatch.Provider))
	case errors.As(err, &unreachable):
		diags.AddError("Project not reachable",
			fmt.Sprintf("This resource is in project %s, but the provider now targets project %s with a project API key, which only reaches its own project. Nothing was changed. If the API key or project_id changed by mistake, restore it. To stop managing the resource, remove it from state (terraform state rm, or a removed block). To move it to project %s, use an organization API key with access to both projects; the resource is then replaced.",
				unreachable.Stored, unreachable.Target, unreachable.Target))
	case errors.As(err, &access):
		diags.AddError("Project not accessible", err.Error())
	default:
		diags.AddError("Error resolving project", err.Error())
	}
	return diags
}

type projectScoped struct {
	client sdkclient.Client
}

func (s *projectScoped) scope() projectscope.Scope {
	return s.client.Scope
}

// clientForTarget returns a client for the project a create or a data
// source read targets. configured is the project_id from the plan or the
// configuration.
func (s *projectScoped) clientForTarget(configured types.String, diags *diag.Diagnostics) (*sdkclient.Client, bool) {
	target, err := s.scope().Target(configured.ValueString())
	if err != nil {
		diags.Append(ProjectDiagnostics(err)...)
		return nil, false
	}
	client := s.client
	if target != "" {
		client = client.WithProject(target, targetProjectHint)
	}
	return &client, true
}

// ProjectScopedResource is embedded by resources that live in a project. It
// provides Configure and ModifyPlan, and the clients CRUD uses.
type ProjectScopedResource struct {
	projectScoped

	// ParentAttribute names the attribute holding the id of the object this
	// resource is part of, for a resource with no identity of its own. Such
	// a resource is in its parent's project: it moves with a new parent and
	// is never replaced to change project.
	ParentAttribute string
}

func (r *ProjectScopedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

// SingleProject reports whether the provider manages one project.
func (r *ProjectScopedResource) SingleProject() bool {
	return r.scope().SingleProject()
}

// ClientForCreate returns a client for the project in the plan.
func (r *ProjectScopedResource) ClientForCreate(planned types.String, diags *diag.Diagnostics) (*sdkclient.Client, bool) {
	return r.clientForTarget(planned, diags)
}

// ClientForState returns a client for the project the state records, so
// read, update and delete stay in the resource's project whatever the
// provider configuration says now.
func (r *ProjectScopedResource) ClientForState(stored string, diags *diag.Diagnostics) (*sdkclient.Client, bool) {
	if stored == "" {
		if !r.scope().SingleProject() {
			diags.AddError("Missing project_id",
				"The state of this resource does not record its project, and the provider is configured with an organization API key and no project_id. Set project_id on the provider for this run, or re-import the resource as <project_id>/<id>.")
			return nil, false
		}
		return r.clientForTarget(types.StringNull(), diags)
	}
	if err := r.scope().Reach(stored); err != nil {
		diags.Append(ProjectDiagnostics(err)...)
		return nil, false
	}
	client := r.client.WithProject(stored, storedProjectHint)
	return &client, true
}

// ModifyPlan checks the resource's project at plan time: a project_id the
// provider mode does not allow is an error, and a change of project is a
// replace.
func (r *ProjectScopedResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var configured types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(projectIDAttribute), &configured)...)
	stored := ""
	if !req.State.Raw.IsNull() {
		stored = storedProject(ctx, req.State, &resp.Diagnostics)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if configured.IsUnknown() {
		if stored != "" && r.ParentAttribute == "" {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root(projectIDAttribute))
		}
		return
	}

	target, action, err := r.scope().Plan(configured.ValueString(), stored)
	if err != nil {
		resp.Diagnostics.Append(ProjectDiagnostics(err)...)
		return
	}

	if action == projectscope.ActionReplace && r.ParentAttribute != "" {
		r.planParentProject(ctx, req, resp, stored, target)
		return
	}
	if action == projectscope.ActionReplace {
		if err := r.client.CheckProject(ctx, target, targetProjectHint); err != nil {
			resp.Diagnostics.Append(ProjectDiagnostics(err)...)
			return
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(projectIDAttribute), target)...)
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root(projectIDAttribute))
		return
	}

	var planned types.String
	resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root(projectIDAttribute), &planned)...)
	if planned.IsUnknown() && target != "" {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(projectIDAttribute), target)...)
	}
}

func (r *ProjectScopedResource) planParentProject(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, stored, target string) {
	var planParent, stateParent types.String
	resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root(r.ParentAttribute), &planParent)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(r.ParentAttribute), &stateParent)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if planParent.Equal(stateParent) {
		resp.Diagnostics.AddAttributeError(path.Root(projectIDAttribute), "Project mismatch",
			fmt.Sprintf("This resource belongs to the project of %s %s, which is %s, but its project resolves to %s. Nothing was changed. Point project_id (or the provider project_id) back at %s, or change %s.",
				r.ParentAttribute, stateParent.ValueString(), stored, target, stored, r.ParentAttribute))
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(projectIDAttribute), target)...)
}

func storedProject(ctx context.Context, state tfsdk.State, diags *diag.Diagnostics) string {
	var projectID, teamID types.String
	diags.Append(state.GetAttribute(ctx, path.Root(projectIDAttribute), &projectID)...)
	if _, d := state.Schema.AttributeAtPath(ctx, path.Root(teamIDAttribute)); !d.HasError() {
		diags.Append(state.GetAttribute(ctx, path.Root(teamIDAttribute), &teamID)...)
	}
	return StoredProject(projectID, teamID)
}

// ImportProjectState accepts "<id>" or "<project_id>/<id>" and writes the id
// to idAttribute and, when present, the project to project_id.
func (r *ProjectScopedResource) ImportProjectState(ctx context.Context, idAttribute string, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID, id, err := projectscope.ParseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", err.Error())
		return
	}
	if projectID == "" && !r.scope().SingleProject() {
		resp.Diagnostics.AddError("Missing project in import ID",
			"The provider is configured with an organization API key and no project_id. Import the resource as <project_id>/<id>.")
		return
	}
	if _, err := r.scope().Target(projectID); err != nil {
		resp.Diagnostics.Append(ProjectDiagnostics(err)...)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(idAttribute), id)...)
	if projectID != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(projectIDAttribute), projectID)...)
	}
}

// ProjectScopedDataSource is embedded by data sources that read from a
// project.
type ProjectScopedDataSource struct {
	projectScoped
}

func (d *ProjectScopedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

// Client returns a client for the project the data source reads from.
func (d *ProjectScopedDataSource) Client(configured types.String, diags *diag.Diagnostics) (*sdkclient.Client, bool) {
	return d.clientForTarget(configured, diags)
}
