package provider

import (
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/net/context"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ resource.Resource              = &secretResource{}
	_ resource.ResourceWithConfigure = &secretResource{}
)

// NewSecretResource is a helper function to simplify the provider implementation.
func NewSecretResource() resource.Resource {
	return &secretResource{}
}

// secretResource defines the data source implementation.
type secretResource struct {
	bitwardenClient sdk.BitwardenClientInterface
	organizationId  string
}

type secretResourceModel struct {
	ID             types.String            `tfsdk:"id"`
	Key            types.String            `tfsdk:"key"`
	Value          types.String            `tfsdk:"value"`
	Note           types.String            `tfsdk:"note"`
	ProjectIDs     []secretProjectIdsModel `tfsdk:"project_ids"`
	OrganizationID types.String            `tfsdk:"organization_id"`
	CreationDate   types.String            `tfsdk:"creation_date"`
	RevisionDate   types.String            `tfsdk:"revision_date"`
}

type secretProjectIdsModel struct {
	ProjectID types.String `tfsdk:"project_id"`
}

func (s *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the content of a secret.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The id of the secret.",
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "The key of the secret.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the secret.",
				Required:    true,
				Sensitive:   true,
			},
			"note": schema.StringAttribute{
				Description: "The note of the secret.",
				Computed:    true,
			},
			"project_ids": schema.ListNestedAttribute{
				Description: "The project ids of the secret.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Description: "One project id of the secret.",
							Computed:    true,
						},
					},
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization id of the secret.",
				Required:    true,
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

func (s *secretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling BitwardenSecretsManagerProviderDataStruct because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring Secret Resource")
	if req.ProviderData == nil {
		tflog.Debug(ctx, "Skipping Resource Configuration because Provider has not been configured yet.")
		return
	}

	providerDataStruct, ok := req.ProviderData.(BitwardenSecretsManagerProviderDataStruct)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
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

	tflog.Info(ctx, "Resource Configured")
}

func (s *secretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
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

	var projectIDs []string
	for _, projectID := range plan.ProjectIDs {
		projectIDs = append(projectIDs, projectID.ProjectID.ValueString())
	}

	secret, err := s.bitwardenClient.Secrets().Create(
		plan.Key.ValueString(),
		plan.Value.ValueString(),
		plan.Note.ValueString(),
		plan.OrganizationID.ValueString(),
		projectIDs,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Secret",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(secret.ID)
	plan.CreationDate = types.StringValue(secret.CreationDate)
	plan.RevisionDate = types.StringValue(secret.RevisionDate)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (s *secretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading Secret Resource")

	var state secretResourceModel
	diags := req.State.Get(ctx, &state)
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
	state.ProjectIDs = []secretProjectIdsModel{{ProjectID: types.StringValue(*secret.ProjectID)}}
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

func (s *secretResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (s *secretResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}
