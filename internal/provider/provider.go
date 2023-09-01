package provider

import (
	"context"
	"os"

	"terraform-provider-hookdeck/internal/provider/connection"
	"terraform-provider-hookdeck/internal/provider/destination"
	"terraform-provider-hookdeck/internal/provider/source"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
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
	APIBase types.String `tfsdk:"api_base"`
	APIKey  types.String `tfsdk:"api_key"`
}

func (p *hookdeckProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hookdeck"
	resp.Version = p.version
}

func (p *hookdeckProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_base": schema.StringAttribute{
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
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
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HOOKDECK_API_BASE environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Hookdeck API Key",
			"The provider cannot create the Hookdeck API client as there is an unknown configuration value for the Hookdeck API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HOOKDECK_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiBase := os.Getenv("HOOKDECK_API_BASE")
	apiKey := os.Getenv("HOOKDECK_API_KEY")

	if !config.APIBase.IsNull() {
		apiBase = config.APIBase.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiBase == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_base"),
			"Missing Hookdeck API base URL",
			"The provider cannot create the Hookdeck API client as there is a missing or empty value for the Hookdeck API base URL. "+
				"Set the base URL value in the configuration or use the HOOKDECK_API_BASE environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Hookdeck API Key",
			"The provider cannot create the Hookdeck API client as there is a missing or empty value for the Hookdeck API key. "+
				"Set the api key value in the configuration or use the HOOKDECK_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "hookdeck_api_base", apiBase)
	ctx = tflog.SetField(ctx, "hookdeck_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hookdeck_api_key")

	tflog.Debug(ctx, "Creating Hookdeck client")
	tflog.Debug(ctx, apiBase+" "+apiKey)

	// Create a new Hookdeck client using the configuration values
	client := hookdeckClient.NewClient(
		hookdeckClient.ClientWithBaseURL(apiBase),
		hookdeckClient.ClientWithAuthToken(apiKey),
	)

	// Make the Hookdeck client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Hookdeck client", map[string]any{"success": true})
}

func (p *hookdeckProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		connection.NewConnectionResource,
		destination.NewDestinationResource,
		source.NewSourceResource,
	}
}

func (p *hookdeckProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &hookdeckProvider{
			version: version,
		}
	}
}
