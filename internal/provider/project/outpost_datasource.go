package project

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-hookdeck/internal/provider/shared"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource                     = &outpostProjectDataSource{}
	_ datasource.DataSourceWithConfigure        = &outpostProjectDataSource{}
	_ datasource.DataSourceWithConfigValidators = &outpostProjectDataSource{}
)

// NewOutpostProjectDataSource returns the hookdeck_outpost_project data source.
func NewOutpostProjectDataSource() datasource.DataSource {
	return &outpostProjectDataSource{}
}

type outpostProjectDataSource struct {
	client sdkclient.Client
}

func (d *outpostProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = outpost.typeName(req.ProviderTypeName)
}

func (d *outpostProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := dataSourceAttributes(outpost)
	attributes["config"] = schema.MapAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "The project's configuration, defaults included, keyed by Outpost environment variable name: every key that has a value and is known to be non-sensitive, whether or not a `hookdeck_outpost_project` resource manages it. A value that is a URL with a user or password is in `sensitive_config` instead.",
	}
	attributes["sensitive_config"] = schema.MapAttribute{
		Computed:    true,
		Sensitive:   true,
		ElementType: types.StringType,
		Description: "The rest of the project's configuration: credential keys, and any key this provider version does not know to be non-sensitive, including keys added to Outpost after this provider release. It includes the project's credentials and is stored in the Terraform state.",
	}
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("%s project, looked up by ID or name.", outpost.label),
		Attributes:  attributes,
	}
}

func (d *outpostProjectDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
	}
}

func (d *outpostProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, diags := shared.ClientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *outpostProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data outpostProjectDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(lookup(ctx, outpost, d.client, &data.projectResourceModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	live, err := getConfig(ctx, d.client.Outpost().WithProject(id, dataSourceKeyHint))
	if sdkclient.IsCode(err, http.StatusConflict, codeNotProvisioned) {
		resp.Diagnostics.AddError("Outpost project is not ready yet", "Project "+id+" is still being provisioned, so its configuration cannot be read. Try again in a few minutes.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading Outpost project configuration", err.Error())
		return
	}
	plain, sensitive := splitConfig(live)
	var diags diag.Diagnostics
	data.Config, diags = types.MapValueFrom(ctx, types.StringType, plain)
	resp.Diagnostics.Append(diags...)
	data.SensitiveConfig, diags = types.MapValueFrom(ctx, types.StringType, sensitive)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
