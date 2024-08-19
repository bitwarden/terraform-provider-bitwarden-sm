package provider

import (
	"context"
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ datasource.DataSource              = &listSecretsDataSource{}
	_ datasource.DataSourceWithConfigure = &listSecretsDataSource{}
)

func NewListSecretsDataSource() datasource.DataSource {
	return &listSecretsDataSource{}
}

// listSecretsDataSource defines the data source implementation.
type listSecretsDataSource struct {
	bitwardenClient sdk.BitwardenClientInterface
	organizationId  string
}

type listSecretsDataSourceModel struct {
	Secrets []listSecretDataSourceModel `tfsdk:"secrets"`
}

type listSecretDataSourceModel struct {
	ID  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

func (l *listSecretsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list_secrets"
}

func (l *listSecretsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "The list_secrets data source fetches all secrets accessible by the used machine account.",
		MarkdownDescription: "The `list_secrets` data source fetches all secrets accessible by the used machine account.",
		Attributes: map[string]schema.Attribute{
			"secrets": schema.ListNestedAttribute{
				Description: "Nested list of all fetched secrets",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:         "String representation of the ID of the secret inside Bitwarden Secrets Manager.",
							MarkdownDescription: "String representation of the `ID` of the secret inside Bitwarden Secrets Manager.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							Description:         "String representation of the key of the secret. Inside Bitwarden Secrets Manager this is called \"name\".",
							MarkdownDescription: "String representation of the `key` of the secret. Inside Bitwarden Secrets Manager this is called \"name\".",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (l *listSecretsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling BitwardenSecretsManagerProviderDataStruct because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring List Secrets Datasource")
	if req.ProviderData == nil {
		tflog.Debug(ctx, "Skipping Datasource Configuration because Provider has not been configured yet.")
		return
	}

	providerDataStruct, ok := req.ProviderData.(BitwardenSecretsManagerProviderDataStruct)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.BitwardenClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	// TODO Due to concurrency issues in an underlying library, we replaced the implementation of a shared client with the initialization of one BitwardenClient per Terraform object
	// Reference: https://github.com/bitwarden/sdk/pull/955
	//client := providerDataStruct.bitwardenClient
	//organizationId := providerDataStruct.organizationId
	//
	//if client == nil {
	//	resp.Diagnostics.AddError(
	//		"Client Not Initialized",
	//		"The Bitwarden bitwardenClient was not properly initialized due to a missing Bitwarden API Client.",
	//	)
	//	return
	//}
	//
	//if organizationId == "" {
	//	resp.Diagnostics.AddError(
	//		"Client Not Initialized",
	//		"The Bitwarden bitwardenClient was not properly initialized due to an empty Organization ID.",
	//	)
	//	return
	//}

	tflog.Debug(ctx, "Creating Bitwarden Secrets Manager Client for List Secrets Datasource")

	apiUrl := providerDataStruct.apiUrl
	if apiUrl == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to an empty API URL.",
		)
		return
	}

	identityUrl := providerDataStruct.identityUrl
	if identityUrl == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to an empty IDENTITY URL.",
		)
		return
	}

	accessToken := providerDataStruct.accessToken
	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to an empty Access Token.",
		)
		return
	}

	organizationId := providerDataStruct.organizationId
	if organizationId == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to an empty Organization ID.",
		)
		return
	}

	// Create a new bitwardenClient using the configuration values
	client, err := sdk.NewBitwardenClient(&apiUrl, &identityUrl)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Bitwarden Secrets Manager Client for List Secrets Datasource",
			"An unexpected error occurred when creating the Bitwarden Secrets Manager Client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bitwarden Secrets Manager Client Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Bitwarden Secrets Manager Client created")

	err = client.AccessTokenLogin(accessToken, &statePath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate Bitwarden Secrets Manager Client for List Secrets Datasource",
			"An unexpected error occurred when authenticating the Bitwarden Secrets Manager Client against the configured endpoint. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bitwarden Secrets Manager Client Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Bitwarden Secrets Manager Client authenticated")

	l.bitwardenClient = client
	l.organizationId = organizationId

	tflog.Info(ctx, "Datasource Configured")
}

func (l *listSecretsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading List Secrets Datasource")

	var state listSecretsDataSourceModel

	if l.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized.",
		)
		return
	}

	secrets, err := l.bitwardenClient.Secrets().List(l.organizationId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Secrets",
			err.Error(),
		)
		return
	}

	for _, secret := range secrets.Data {
		secretState := listSecretDataSourceModel{
			ID:  types.StringValue(secret.ID),
			Key: types.StringValue(secret.Key),
		}
		state.Secrets = append(state.Secrets, secretState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
