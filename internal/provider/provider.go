// Copyright (c) HashiCorp, Inc.

package provider

import (
	"context"
	"log/slog"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/weisshorn-cyd/gocti"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &openctiProvider{}
)

// openctiProviderModel maps provider schema data to a Go type.
type openctiProviderModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &openctiProvider{
			version: version,
		}
	}
}

// openctiProvider is the provider implementation.
type openctiProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *openctiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "opencti"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *openctiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a opencti API client for data sources and resources.
func (p *openctiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring opencti client")

	// Retrieve provider data from configuration
	var config openctiProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown opencti URL",
			"The provider cannot create the opencti API client as there is an unknown configuration value for the opencti URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OPENCTI_URL environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown opencti token",
			"The provider cannot create the opencti API client as there is an unknown configuration value for the opencti token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OPENCTI_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv("OPENCTI_URL")
	token := os.Getenv("OPENCTI_TOKEN")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing opencti URL",
			"The provider cannot create the opencti API client as there is a missing or empty value for the opencti URL. "+
				"Set the url value in the configuration or use the OPENCTI_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token "),
			"Missing opencti token",
			"The provider cannot create the opencti API client as there is a missing or empty value for the opencti token. "+
				"Set the token value in the configuration or use the OPENCTI_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "opencti_url", url)
	ctx = tflog.SetField(ctx, "opencti_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "opencti_token")

	tflog.Debug(ctx, "Creating opencti client")

	// Create a new opencti client using the configuration values
	client, err := gocti.NewOpenCTIAPIClient(
		url,
		token,
		gocti.WithHealthCheck(),
		gocti.WithLogLevel(slog.LevelInfo),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create opencti API Client",
			"An unexpected error occurred when creating the opencti API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"opencti Client Error: "+err.Error(),
		)

		return
	}

	// Make the opencti client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured opencti client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *openctiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *openctiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCaseTemplateResource,
		NewGroupResource,
		NewMarkingDefinitionResource,
		NewRoleResource,
		NewStatusTemplateResource,
		NewTaskTemplateResource,
		NewUserResource,
		NewVocabularyResource,
	}
}
