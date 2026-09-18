package provider

import (
	"context"
	"fmt"
	"os"

	"terraform-provider-hookdeck/internal/projectscope"
	"terraform-provider-hookdeck/internal/provider/connection"
	"terraform-provider-hookdeck/internal/provider/destination"
	"terraform-provider-hookdeck/internal/provider/project"
	"terraform-provider-hookdeck/internal/provider/source"
	"terraform-provider-hookdeck/internal/provider/sourceauth"
	"terraform-provider-hookdeck/internal/provider/transformation"
	"terraform-provider-hookdeck/internal/provider/webhookregistration"
	"terraform-provider-hookdeck/internal/sdkclient"
	"terraform-provider-hookdeck/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies various provider interfaces.
var _ provider.Provider = &hookdeckProvider{}

// hookdeckProvider defines the provider implementation.
type hookdeckProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// hookdeckProviderModel describes the provider data model.
type hookdeckProviderModel struct {
	APIBase   types.String `tfsdk:"api_base"`
	APIKey    types.String `tfsdk:"api_key"`
	ProjectID types.String `tfsdk:"project_id"`
}

func (p *hookdeckProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hookdeck"
	resp.Version = p.version
}

func (p *hookdeckProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_base": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Hookdeck API Base URL. Alternatively, can be configured using the `%s` environment variable.", apiBaseEnvVarKey),
				Validators: []validator.String{
					validators.NewHookdeckAPIBaseURLValidator(),
				},
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: fmt.Sprintf("Hookdeck API Key, either a project key or an organization key (`hd_org_` prefix). Alternatively, can be configured using the `%s` environment variable.", apiKeyEnvVarKey),
			},
			"project_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("Default project for every resource that does not set its own `project_id`. Required with an organization API key unless each resource sets `project_id`. With a project API key it must match the key's project. Alternatively, can be configured using the `%s` environment variable.", projectIDEnvVarKey),
			},
		},
	}
}

func (p *hookdeckProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Hookdeck client")

	// Retrieve provider data from configuration
	var config hookdeckProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.APIBase.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_base"),
			"Unknown Hookdeck API Base URL",
			"The provider cannot create the Hookdeck API client as there is an unknown configuration value for the Hookdeck API base URL. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", apiBaseEnvVarKey),
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Hookdeck API Key",
			"The provider cannot create the Hookdeck API client as there is an unknown configuration value for the Hookdeck API key. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", apiKeyEnvVarKey),
		)
	}

	if config.ProjectID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("project_id"),
			"Unknown Hookdeck Project ID",
			"The provider cannot create the Hookdeck API client as there is an unknown configuration value for the Hookdeck project ID. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", projectIDEnvVarKey),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiBase := os.Getenv(apiBaseEnvVarKey)
	apiKey := os.Getenv(apiKeyEnvVarKey)
	projectID := os.Getenv(projectIDEnvVarKey)

	if !config.APIBase.IsNull() {
		apiBase = config.APIBase.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.ProjectID.IsNull() {
		projectID = config.ProjectID.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Hookdeck API Key",
			"The provider cannot create the Hookdeck API client as there is a missing or empty value for the Hookdeck API key. "+
				fmt.Sprintf("Set the api key value in the configuration or use the %s environment variable. ", apiKeyEnvVarKey)+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "hookdeck_api_base", apiBase)
	ctx = tflog.SetField(ctx, "hookdeck_project_id", projectID)

	tflog.Debug(ctx, "Creating Hookdeck client")

	// Create a new Hookdeck client using the configuration values
	client := sdkclient.InitHookdeckSDKClient(apiBase, apiKey, p.version)
	client.DefaultProjectID = projectID

	// A project key with a provider default that is not its own project is
	// a configuration error; fail here rather than on the first resource.
	if projectID != "" && client.KeyKind == projectscope.KeyKindProject {
		if _, err := client.ForProject(ctx, ""); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("project_id"), "Project mismatch", err.Error())
			return
		}
	}

	// Make the Hookdeck client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hookdeck client", map[string]any{"success": true})
}

func (p *hookdeckProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Event Gateway
		project.NewProjectResource,
		connection.NewConnectionResource,
		destination.NewDestinationResource,
		source.NewSourceResource,
		sourceauth.NewSourceAuthResource,
		transformation.NewTransformationResource,
		webhookregistration.NewWebhookRegistrationResource,
		// v2 names, deprecated, removed in v4
		connection.NewLegacyConnectionResource,
		destination.NewLegacyDestinationResource,
		source.NewLegacySourceResource,
		sourceauth.NewLegacySourceAuthResource,
		transformation.NewLegacyTransformationResource,
	}
}

func (p *hookdeckProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Event Gateway
		project.NewProjectDataSource,
		connection.NewConnectionDataSource,
		destination.NewDestinationDataSource,
		source.NewSourceDataSource,
		// v2 names, deprecated, removed in v4
		connection.NewLegacyConnectionDataSource,
		destination.NewLegacyDestinationDataSource,
		source.NewLegacySourceDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hookdeckProvider{
			version: version,
		}
	}
}

const (
	apiBaseEnvVarKey   = "HOOKDECK_API_BASE"
	apiKeyEnvVarKey    = "HOOKDECK_API_KEY"
	projectIDEnvVarKey = "HOOKDECK_PROJECT_ID"
)
