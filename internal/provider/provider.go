package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"terraform-provider-hookdeck/internal/provider/connection"
	"terraform-provider-hookdeck/internal/provider/destination"
	"terraform-provider-hookdeck/internal/provider/source"
	"terraform-provider-hookdeck/internal/provider/sourceauth"
	"terraform-provider-hookdeck/internal/provider/transformation"
	"terraform-provider-hookdeck/internal/provider/webhookregistration"
	"terraform-provider-hookdeck/internal/sdkclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	APIBase        types.String `tfsdk:"api_base"`
	APIKey         types.String `tfsdk:"api_key"`
	SDKMaxAttempts types.Int32  `tfsdk:"sdk_max_attempts"`
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
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: fmt.Sprintf("Hookdeck API Key. Alternatively, can be configured using the `%s` environment variable.", apiKeyEnvVarKey),
			},
			"sdk_max_attempts": schema.Int32Attribute{
				Optional:            true,
				MarkdownDescription: fmt.Sprintf("The maximum number of attempts to make when sending a request to the Hookdeck API. Alternatively, can be configured using the `%s` environment variable.", sdkMaxAttemptsEnvVarKey),
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

	if config.SDKMaxAttempts.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("sdk_max_attempts"),
			"Unknown Hookdeck SDK Max Attempts",
			"The provider cannot create the Hookdeck API client as there is an unknown configuration value for the SDK max attempts option. "+
				fmt.Sprintf("Either target apply the source of the value first, set the value statically in the configuration, or use the %s environment variable.", sdkMaxAttemptsEnvVarKey),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiBase := os.Getenv(apiBaseEnvVarKey)
	apiKey := os.Getenv(apiKeyEnvVarKey)
	sdkMaxAttempts := defaultSDKMaxAttempts
	sdkMaxAttemptsStr := os.Getenv(sdkMaxAttemptsEnvVarKey)
	if sdkMaxAttemptsStr != "" {
		var err error
		sdkMaxAttempts, err = strconv.Atoi(sdkMaxAttemptsStr)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("sdk_max_attempts"),
				"Invalid Hookdeck SDK Max Attempts",
				"The provider cannot create the Hookdeck API client as there is an invalid configuration value for the SDK max attempts option. ",
			)
		}
	}

	if !config.APIBase.IsNull() {
		apiBase = config.APIBase.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.SDKMaxAttempts.IsNull() {
		sdkMaxAttempts = int(config.SDKMaxAttempts.ValueInt32())
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
	ctx = tflog.SetField(ctx, "hookdeck_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "hookdeck_api_key")
	ctx = tflog.SetField(ctx, "sdk_max_attempts", sdkMaxAttempts)

	tflog.Debug(ctx, "Creating Hookdeck client")
	tflog.Debug(ctx, apiBase+" "+apiKey+" "+strconv.Itoa(sdkMaxAttempts))

	// Create a new Hookdeck client using the configuration values
	client := sdkclient.InitHookdeckSDKClient(apiBase, apiKey, sdkMaxAttempts, p.version)

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
		sourceauth.NewSourceAuthResource,
		transformation.NewTransformationResource,
		webhookregistration.NewWebhookRegistrationResource,
	}
}

func (p *hookdeckProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		connection.NewConnectionDataSource,
		destination.NewDestinationDataSource,
		source.NewSourceDataSource,
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
	apiBaseEnvVarKey        = "HOOKDECK_API_BASE"
	apiKeyEnvVarKey         = "HOOKDECK_API_KEY"
	sdkMaxAttemptsEnvVarKey = "HOOKDECK_SDK_MAX_ATTEMPTS"
)

const (
	defaultSDKMaxAttempts = 1
)
