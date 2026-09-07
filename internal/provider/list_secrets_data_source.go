package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/bitwarden/sdk-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	ProjectID types.String                `tfsdk:"project_id"`
	Filter    types.String                `tfsdk:"filter"`
	Secrets   []listSecretDataSourceModel `tfsdk:"secrets"`
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
		Description:         "The list_secrets data source fetches all secrets accessible by the used machine account. Secrets can optionally be filtered by project ID or by key (name).",
		MarkdownDescription: "The `list_secrets` data source fetches all secrets accessible by the used machine account. Secrets can optionally be filtered by `project_id` or by key (name) using `filter`.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description:         "Filter secrets by the project ID they belong to. Only secrets associated with this project will be returned.",
				MarkdownDescription: "Filter secrets by the `project_id` they belong to. Only secrets associated with this project will be returned.",
				Optional:            true,
				Validators:          []validator.String{stringUUIDValidate()},
			},
			"filter": schema.StringAttribute{
				Description:         "Filter secrets by key (name). Only secrets whose key contains this value (case-insensitive) will be returned.",
				MarkdownDescription: "Filter secrets by `key` (name). Only secrets whose key contains this value (case-insensitive) will be returned.",
				Optional:            true,
			},
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

	client := providerDataStruct.bitwardenClient
	organizationId := providerDataStruct.organizationId

	if client == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to a missing Bitwarden API Client.",
		)
		return
	}

	if organizationId == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized due to an empty Organization ID.",
		)
		return
	}

	l.bitwardenClient = client
	l.organizationId = organizationId

	tflog.Info(ctx, "Datasource Configured")
}

func (l *listSecretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading List Secrets Datasource")

	var config listSecretsDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	filterByProjectID := !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown()
	filterByKey := !config.Filter.IsNull() && !config.Filter.IsUnknown()

	var state listSecretsDataSourceModel
	state.ProjectID = config.ProjectID
	state.Filter = config.Filter

	// When filtering by project_id, we need to fetch full secret details
	// because the list API only returns secret identifiers without project info.
	if filterByProjectID && len(secrets.Data) > 0 {
		projectIDFilter := config.ProjectID.ValueString()

		var secretIDs []string
		for _, secret := range secrets.Data {
			secretIDs = append(secretIDs, secret.ID)
		}

		fullSecrets, err := l.bitwardenClient.Secrets().GetByIDS(secretIDs)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Fetch Secret Details",
				err.Error(),
			)
			return
		}

		// Replace the list data with only the secrets matching the project filter.
		secrets.Data = nil
		for _, secret := range fullSecrets.Data {
			if secret.ProjectID != nil && *secret.ProjectID == projectIDFilter {
				secrets.Data = append(secrets.Data, sdk.SecretIdentifierResponse{
					ID:             secret.ID,
					Key:            secret.Key,
					OrganizationID: secret.OrganizationID,
				})
			}
		}
	}

	keyFilter := ""
	if filterByKey {
		keyFilter = strings.ToLower(config.Filter.ValueString())
	}

	for _, secret := range secrets.Data {
		if keyFilter != "" && !strings.Contains(strings.ToLower(secret.Key), keyFilter) {
			continue
		}
		state.Secrets = append(state.Secrets, listSecretDataSourceModel{
			ID:  types.StringValue(secret.ID),
			Key: types.StringValue(secret.Key),
		})
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
