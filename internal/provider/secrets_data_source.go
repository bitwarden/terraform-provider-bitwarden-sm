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
	_ datasource.DataSource              = &secretsDataSource{}
	_ datasource.DataSourceWithConfigure = &secretsDataSource{}
)

func NewSecretsDataSource() datasource.DataSource {
	return &secretsDataSource{}
}

// secretsDataSource defines the data source implementation.
type secretsDataSource struct {
	bitwardenClient *sdk.BitwardenClientInterface
}

type secretsDataSourceModel struct {
	Secrets        []secretDataSourceModel `tfsdk:"secrets"`
	OrganizationId types.String            `tfsdk:"organization_id"`
}

type secretDataSourceModel struct {
	Id  types.String `tfsdk:"id"`
	Key types.String `tfsdk:"key"`
}

func (s *secretsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secrets"
}

func (s *secretsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of secrets accessible by the machine account.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "The identifier of the organization this secrets belongs to.",
				Required:    true,
			},
			"secrets": schema.ListNestedAttribute{
				Description: "List of secrets accessible by the machine account.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Placeholder identifier attribute.",
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "The key of the secret.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (s *secretsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring Secrets Datasource")
	if req.ProviderData == nil {
		tflog.Debug(ctx, "Skipping Datasource Configuration because Provider has not been configured yet.")
		return
	}

	client, ok := req.ProviderData.(*sdk.BitwardenClientInterface)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sdk.BitwardenClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	if client == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized.",
		)
		return
	}

	s.bitwardenClient = client

	tflog.Info(ctx, "Datasource Configured")
}

func (s *secretsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Info(ctx, "Reading Secrets Datasource")

	var state secretsDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if s.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden bitwardenClient was not properly initialized.",
		)
		return
	}

	client := *s.bitwardenClient
	secrets, err := client.Secrets().List(state.OrganizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Secrets",
			err.Error(),
		)
		return
	}

	for _, secret := range secrets.Data {
		secretState := secretDataSourceModel{
			Id:  types.StringValue(secret.ID),
			Key: types.StringValue(secret.Key),
		}
		state.Secrets = append(state.Secrets, secretState)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
