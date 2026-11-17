package project

import (
	"context"
	"fmt"

	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/schemahelpers"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource                     = &projectDataSource{}
	_ datasource.DataSourceWithConfigure        = &projectDataSource{}
	_ datasource.DataSourceWithConfigValidators = &projectDataSource{}
)

// NewProjectDataSource returns the hookdeck_gateway_project data source.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{kind: gateway}
}

type projectDataSource struct {
	kind   kind
	client sdkclient.Client
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = d.kind.typeName(req.ProviderTypeName)
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("%s project, looked up by ID or name.", d.kind.label),
		Attributes:  dataSourceAttributes(d.kind),
	}
}

// dataSourceAttributes are the attributes every project data source has.
func dataSourceAttributes(k kind) map[string]schema.Attribute {
	attributes := schemahelpers.DataSourceSchemaFromResourceSchema(schemaAttributes(k), "id")
	attributes["id"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "ID of the project. One of `id` or `name` is required.",
	}
	attributes["name"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: fmt.Sprintf("Name of the project. One of `id` or `name` is required; the name must be unique among the %s projects the API key can see.", k.label),
	}
	return attributes
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
	resp.Diagnostics.Append(lookup(ctx, d.kind, d.client, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// lookup fills data from the project its id or name identifies.
func lookup(ctx context.Context, k kind, client sdkclient.Client, data *projectResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if !data.ID.IsNull() {
		found, d := data.retrieve(ctx, k, client)
		diags.Append(d...)
		if !found && !diags.HasError() {
			diags.AddError("Project not found", "No project with ID "+data.ID.ValueString()+" is visible to this API key.")
		}
		return diags
	}
	project, d := findByName(ctx, k, client, data.Name.ValueString())
	diags.Append(d...)
	if !diags.HasError() {
		diags.Append(data.refresh(k, project)...)
	}
	return diags
}
