package provider

import (
	"context"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
)

var (
	// Ensure BitwardenSecretsManagerProvider satisfies various provider interfaces.
	_         provider.Provider              = &BitwardenSecretsManagerProvider{}
	_         provider.ProviderWithFunctions = &BitwardenSecretsManagerProvider{}
	statePath                                = ".bw-provider-state"
)

// BitwardenSecretsManagerProvider defines the provider implementation.
type BitwardenSecretsManagerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BitwardenSecretsManagerProviderModel describes the provider data model.
type BitwardenSecretsManagerProviderModel struct {
	ApiUrl         types.String `tfsdk:"api_url"`
	IdentityUrl    types.String `tfsdk:"identity_url"`
	AccessToken    types.String `tfsdk:"access_token"`
	OrganizationId types.String `tfsdk:"organization_id"`
}

func (p *BitwardenSecretsManagerProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bitwarden-sm"
	resp.Version = p.version
}

type BitwardenSecretsManagerProviderDataStruct struct {
	bitwardenClient sdk.BitwardenClientInterface
	organizationId  string
}

func (p *BitwardenSecretsManagerProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This Terraform provider interacts with Bitwarden Secrets Manager to manage Secrets and Projects.",
		MarkdownDescription: "This Terraform provider interacts with [**Bitwarden Secrets Manager**](https://bitwarden.com/products/secrets-manager/) " +
			"to manage `Secrets` and `Projects`.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "URI for the Bitwarden Secrets Manager API endpoint. " +
					"This configuration value is optional because it can also be provided via BW_API_URL environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				MarkdownDescription: "URI for the **Bitwarden Secrets Manager** `API` endpoint. " +
					"This configuration value is _**optional**_ because it can also be provided via `BW_API_URL` environment variable.  " +
					"However, it **must be provided** in one of these two ways.",
				Optional: true,
			},
			"identity_url": schema.StringAttribute{
				Description: "URI for the Bitwarden Secrets Manager IDENTITY endpoint. " +
					"This configuration value is optional because it can also be provided via BW_IDENTITY_API_URL environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				MarkdownDescription: "URI for the **Bitwarden Secrets Manager** `IDENTITY` endpoint. " +
					"This configuration value is _**optional**_ because it can also be provided via `BW_IDENTITY_API_URL` environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				Optional: true,
			},
			"access_token": schema.StringAttribute{
				Description: "Access Token of the used Machine Account for Bitwarden Secrets Manager." +
					"This configuration value is optional because it can also be provided via BW_ACCESS_TOKEN environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				MarkdownDescription: "`Access Token` of the used Machine Account for Bitwarden Secrets Manager. " +
					"This configuration value is _**optional**_ because it can also be provided via `BW_ACCESS_TOKEN` environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				Optional:  true,
				Sensitive: true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of your Organization in Bitwarden Secrets Manager. " +
					"This configuration value is optional because it can also be provided via BW_ORGANIZATION_ID environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				MarkdownDescription: "The `ID` of your Organization in Bitwarden Secrets Manager endpoints. " +
					"This configuration value is _**optional**_ because it can also be provided via `BW_ORGANIZATION_ID` environment variable. " +
					"However, it **must be provided** in one of these two ways.",
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringUUIDValidate(),
				},
			},
		},
	}
}

func (p *BitwardenSecretsManagerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	tflog.Info(ctx, "Configuring Bitwarden Secrets Manager")

	var config BitwardenSecretsManagerProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ApiUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Unknown URI for Bitwarden Secrets Manager API endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is an unknown configuration value for the URI of the Bitwarden Secrets Manager API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BW_API_URL environment variable.",
		)
	}

	if config.IdentityUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("identity_url"),
			"Unknown URI for Bitwarden Secrets Manager IDENTITY endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is an unknown configuration value for the URI of the Bitwarden Secrets Manager IDENTITY endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BW_IDENTITY_API_URL environment variable.",
		)
	}

	if config.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown Access Token for Bitwarden Secrets Manager endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is an unknown configuration value for the Access Token of Bitwarden Secrets Manager endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BW_ACCESS_TOKEN environment variable.",
		)
	}

	if config.OrganizationId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown Organization ID for Bitwarden Secrets Manager endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is an unknown configuration value for the Organization of Bitwarden Secrets Manager endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BW_ORGANIZATION_ID environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	apiUrl := os.Getenv("BW_API_URL")
	identityUrl := os.Getenv("BW_IDENTITY_API_URL")
	accessToken := os.Getenv("BW_ACCESS_TOKEN")
	organizationId := os.Getenv("BW_ORGANIZATION_ID")

	if !config.ApiUrl.IsNull() {
		apiUrl = config.ApiUrl.ValueString()
	}

	if !config.IdentityUrl.IsNull() {
		identityUrl = config.IdentityUrl.ValueString()
	}

	if !config.AccessToken.IsNull() {
		accessToken = config.AccessToken.ValueString()
	}

	if !config.OrganizationId.IsNull() {
		organizationId = config.OrganizationId.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if apiUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Missing URI for Bitwarden Secrets Manager API endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is a missing or empty configuration value for the URI of the Bitwarden Secrets Manager API endpoint. "+
				"Set the api_url value in the configuration or use the BW_API_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if identityUrl == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("identity_url"),
			"Missing URI for Bitwarden Secrets Manager IDENTITY endpoint",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is a missing or empty configuration value for the URI of the Bitwarden Secrets Manager IDENTITY endpoint. "+
				"Set the identity_url value in the configuration or use the BW_IDENTITY_API_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Bitwarden Secrets Manager Access Token",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is a missing or empty configuration value for the Access Token of Bitwarden Secrets Manager endpoint. "+
				"Set the access_token value in the configuration or use the BW_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if organizationId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization_id"),
			"Missing Bitwarden Secrets Manager Organization ID",
			"The provider cannot create the Bitwarden Secrets Manager API bitwardenClient as there is a missing or empty configuration value for the Organization ID of Bitwarden Secrets Manager endpoint. "+
				"Set the organization_id value in the configuration or use the BW_ORGANIZATION_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "bitwarden_secrets_manager_api_url", apiUrl)
	ctx = tflog.SetField(ctx, "bitwarden_secrets_manager_identity_url", identityUrl)
	ctx = tflog.SetField(ctx, "bitwarden_secrets_manager_access_token", accessToken)
	ctx = tflog.SetField(ctx, "bitwarden_secrets_manager_organization_id", organizationId)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "bitwarden_secrets_manager_access_token")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "bitwarden_secrets_manager_organization_id")

	tflog.Debug(ctx, "Creating Bitwarden Secrets Manager Client")

	// Create a new bitwardenClient using the configuration values
	bitwardenClient, err := sdk.NewBitwardenClient(&apiUrl, &identityUrl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Bitwarden Secrets Manager Client",
			"An unexpected error occurred when creating the Bitwarden Secrets Manager Client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bitwarden Secrets Manager Client Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Bitwarden Secrets Manager Client created")

	err = bitwardenClient.AccessTokenLogin(accessToken, &statePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate Bitwarden Secrets Manager Client",
			"An unexpected error occurred when authenticating the Bitwarden Secrets Manager Client against the configured endpoint. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bitwarden Secrets Manager Client Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Bitwarden Secrets Manager Client authenticated")

	// Make the bitwardenClient available during DataSource and Resource
	// type Configure methods.
	providerDataStruct := BitwardenSecretsManagerProviderDataStruct{
		bitwardenClient,
		organizationId,
	}

	resp.DataSourceData = providerDataStruct
	resp.ResourceData = providerDataStruct

	tflog.Info(ctx, "Configured Bitwarden Secrets Manager Client", map[string]any{"success": true})
}

func (p *BitwardenSecretsManagerProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretResource,
	}
}

func (p *BitwardenSecretsManagerProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectsDataSource,
		NewListSecretsDataSource,
		NewSecretDataSource,
	}
}

func (p *BitwardenSecretsManagerProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BitwardenSecretsManagerProvider{
			version: version,
		}
	}
}
