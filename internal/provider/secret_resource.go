package provider

import (
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/net/context"
	"math/rand"
	"time"
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ resource.Resource                = &secretResource{}
	_ resource.ResourceWithConfigure   = &secretResource{}
	_ resource.ResourceWithImportState = &secretResource{}
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
	ID             types.String `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	Value          types.String `tfsdk:"value"`
	Note           types.String `tfsdk:"note"`
	ProjectID      types.String `tfsdk:"project_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

func (s *secretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (s *secretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "The secret resource manages secrets in Bitwarden Secrets Manager.",
		MarkdownDescription: "The `secret` resource manages secrets in Bitwarden Secrets Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "String representation of the ID of the secret inside Bitwarden Secrets Manager.",
				MarkdownDescription: "String representation of the `ID` of the secret inside Bitwarden Secrets Manager.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description:         "String representation of the key of the secret. Inside Bitwarden Secrets Manager this is called \"name\".",
				MarkdownDescription: "String representation of the `key` of the secret. Inside Bitwarden Secrets Manager this is called \"name\".",
				Required:            true,
			},
			"value": schema.StringAttribute{
				Description:         "String representation of the value of the secret inside Bitwarden Secrets Manager. This attribute is sensitive. The Dynamic Secrets feature enables compatibility with secret value changes in Bitwarden Secrets Manager without changes to the terraform plan.",
				MarkdownDescription: "String representation of the `value` of the secret inside Bitwarden Secrets Manager. This attribute is sensitive. The Dynamic Secrets feature enables compatibility with secret `value` changes in Bitwarden Secrets Manager without changes to the terraform plan.",
				Computed:            true,
				Optional:            true,
				Sensitive:           true,
			},
			"note": schema.StringAttribute{
				Description:         "String representation of the note of the secret inside Bitwarden Secrets Manager.",
				MarkdownDescription: "String representation of the `note` of the secret inside Bitwarden Secrets Manager.",
				Computed:            true,
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				Description:         "String representation of the ID of the project to which the secrets belongs. If the used machine account has no read access to this project, access will not be granted.",
				MarkdownDescription: "String representation of the `ID` of the project to which the secret belongs. If the used machine account has no read access to this project, access will not be granted.",
				Computed:            true,
				Optional:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "String representation of the ID of the organization to which the secrets belongs.",
				MarkdownDescription: "String representation of the `ID` of the organization to which the secret belongs.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_date": schema.StringAttribute{
				Description: "String representation of the creation date of the secret.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision_date": schema.StringAttribute{
				Description: "String representation of the revision date of the secret.",
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

	tflog.Debug(ctx, "Creating Bitwarden Secrets Manager Client for Secrets Resource.")

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
			"Unable to Create Bitwarden Secrets Manager Client for Secrets Resource.",
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
			"Unable to Authenticate Bitwarden Secrets Manager Client for Secrets Resource.",
			"An unexpected error occurred when authenticating the Bitwarden Secrets Manager Client against the configured endpoint. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Bitwarden Secrets Manager Client Error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Bitwarden Secrets Manager Client authenticated")

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

	// TODO: Dummy implementation for testing purposes, replace once secretValueCreate(rules... string) is part of go-sdk
	var value string
	if plan.Value.IsUnknown() {
		value = createSecretValue()
	} else {
		value = plan.Value.ValueString()
	}

	secret, err := s.bitwardenClient.Secrets().Create(
		plan.Key.ValueString(),
		value,
		plan.Note.ValueString(),
		s.organizationId,
		[]string{plan.ProjectID.ValueString()},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Secret",
			err.Error(),
		)
		return
	}

	var state secretResourceModel
	state.ID = types.StringValue(secret.ID)
	state.Key = types.StringValue(secret.Key)
	state.Value = types.StringValue(secret.Value)
	state.Note = types.StringValue(secret.Note)
	state.ProjectID = types.StringValue(*secret.ProjectID)
	state.OrganizationID = types.StringValue(secret.OrganizationID)
	state.CreationDate = types.StringValue(secret.CreationDate)
	state.RevisionDate = types.StringValue(secret.RevisionDate)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
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

func (s *secretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan secretResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state secretResourceModel
	diags = req.State.Get(ctx, &state)
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

	key := plan.Key.ValueString()
	if key == "" {
		key = state.Key.ValueString()
	}
	value := plan.Value.ValueString()
	if value == "" {
		value = state.Value.ValueString()
	}
	note := plan.Note.ValueString()
	if note == "" {
		note = state.Note.ValueString()
	}
	projectID := plan.ProjectID.ValueString()
	if projectID == "" {
		projectID = state.ProjectID.ValueString()
	}

	secret, err := s.bitwardenClient.Secrets().Update(
		state.ID.ValueString(),
		key,
		value,
		note,
		state.OrganizationID.ValueString(),
		[]string{projectID},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Secret",
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

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *secretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var plan secretResourceModel
	diags := req.State.Get(ctx, &plan)
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

	secretDeleteResponse, err := s.bitwardenClient.Secrets().Delete([]string{plan.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Secret",
			err.Error(),
		)
		return
	}
	if secretDeleteResponse.Data[0].Error != nil {
		resp.Diagnostics.AddError(
			"Error deleting Secret",
			*secretDeleteResponse.Data[0].Error,
		)
	}
}

func (s *secretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func createSecretValue() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[seed.Intn(len(chars))]
	}

	return string(b)
}
