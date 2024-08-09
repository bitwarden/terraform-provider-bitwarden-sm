package provider

import (
	"context"
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ datasource.DataSource              = &secretDataSource{}
	_ datasource.DataSourceWithConfigure = &secretDataSource{}
)

func NewSecretDataSource() datasource.DataSource {
	return &secretDataSource{}
}

// secretDataSource defines the data source implementation.
type secretDataSource struct {
	bitwardenClient sdk.BitwardenClientInterface
	organizationId  string
}

type secretDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

func (s *secretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the content of a secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the secret.",
				Required:    true,
				Validators: []validator.String{
					stringUUIDValidate(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The key of the secret.",
				Computed:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"note": schema.StringAttribute{
				Description: "The note of the secret.",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project id of the secret.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization id of the secret.",
				Computed:    true,
			},
			"creation_date": schema.StringAttribute{
				Description: "The creation date of the secret.",
				Computed:    true,
			},
			"revision_date": schema.StringAttribute{
				Description: "The revision date of the secret.",
				Computed:    true,
			},
		},
	}
}

func (s *secretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling BitwardenSecretsManagerProviderDataStruct because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring Secret Datasource")
	if req.ProviderData == nil {
		tflog.Debug(ctx, "Skipping Datasource Configuration because Provider has not been configured yet.")
		return
	}

	providerDataStruct, ok := req.ProviderData.(BitwardenSecretsManagerProviderDataStruct)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected sdk.BitwardenClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client := providerDataStruct.bitwardenClient
	organizationId := providerDataStruct.organizationId

	if client == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized due to a missing Bitwarden API Client.",
		)
		return
	}

	if organizationId == "" {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized due to an empty Organization ID.",
		)
		return
	}

	s.bitwardenClient = client
	s.organizationId = organizationId

	tflog.Info(ctx, "Datasource Configured")
}

func (s *secretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading Secret Datasource")

	var state secretDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if s.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	secret, err := s.bitwardenClient.Secrets().Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Secret with id: "+state.ID.ValueString(),
			err.Error(),
		)
		return
	}

	state.Key = types.StringValue(secret.Key)
	state.Value = types.StringValue(secret.Value)
	state.Note = types.StringValue(secret.Note)
	state.ProjectID = types.StringValue(*secret.ProjectID)
	state.OrganizationID = types.StringValue(secret.OrganizationID)
	state.CreationDate = types.StringValue(secret.CreationDate)
	state.RevisionDate = types.StringValue(secret.RevisionDate)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
