package project

import (
	"context"

	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/schemahelpers"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource                     = &projectDataSource{}
	_ datasource.DataSourceWithConfigure        = &projectDataSource{}
	_ datasource.DataSourceWithConfigValidators = &projectDataSource{}
)

// NewProjectDataSource returns the hookdeck_gateway_project data source.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client sdkclient.Client
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(), "id")
	attributes["id"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "ID of the project. One of `id` or `name` is required.",
	}
	attributes["name"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Name of the project. One of `id` or `name` is required; the name must be unique among the projects the API key can see.",
	}
	resp.Schema = schema.Schema{
		Description: "Event Gateway project, looked up by id or name.",
		Attributes:  attributes,
	}
}

func (d *projectDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := shared.ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() {
		found, diags := data.retrieve(ctx, d.client)
		resp.Diagnostics.Append(diags...)
		if !found && !resp.Diagnostics.HasError() {
			resp.Diagnostics.AddError("Project not found", "No project with id "+data.ID.ValueString()+" is visible to this API key.")
		}
	} else {
		project, diags := findByName(ctx, d.client, data.Name.ValueString())
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.Diagnostics.Append(data.refresh(project)...)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
