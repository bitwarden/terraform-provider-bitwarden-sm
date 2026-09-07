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
)

var (
	// Ensure provider defined types fully satisfy framework interfaces.
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewProjectResource is a helper function to simplify the provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource defines the resource implementation.
type projectResource struct {
	bitwardenClient sdk.BitwardenClientInterface
	organizationId  string
}

type projectResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

func (p *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (p *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "The project resource manages projects in Bitwarden Secrets Manager.",
		MarkdownDescription: "The `project` resource manages projects in Bitwarden Secrets Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "String representation of the ID of the project inside Bitwarden Secrets Manager.",
				MarkdownDescription: "String representation of the `ID` of the project inside Bitwarden Secrets Manager.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description:         "String representation of the name of the project inside Bitwarden Secrets Manager.",
				MarkdownDescription: "String representation of the `name` of the project inside Bitwarden Secrets Manager.",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				Description:         "String representation of the ID of the organization to which the project belongs.",
				MarkdownDescription: "String representation of the `ID` of the organization to which the project belongs.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"creation_date": schema.StringAttribute{
				Description:         "String representation of the creation date of the project.",
				MarkdownDescription: "String representation of the `creation_date` of the project.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision_date": schema.StringAttribute{
				Description:         "String representation of the revision date of the project.",
				MarkdownDescription: "String representation of the `revision_date` of the project.",
				Computed:            true,
			},
		},
	}
}

func (p *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling BitwardenSecretsManagerProviderDataStruct because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring Project Resource")
	if req.ProviderData == nil {
		tflog.Debug(ctx, "Skipping Resource Configuration because Provider has not been configured yet.")
		return
	}

	providerDataStruct, ok := req.ProviderData.(BitwardenSecretsManagerProviderDataStruct)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected BitwardenSecretsManagerProviderDataStruct, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	p.bitwardenClient = client
	p.organizationId = organizationId

	tflog.Info(ctx, "Resource Configured")
}

func (p *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if p.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	project, err := p.bitwardenClient.Projects().Create(
		p.organizationId,
		plan.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Project",
			err.Error(),
		)
		return
	}

	var state projectResourceModel
	state.ID = types.StringValue(project.ID)
	state.Name = types.StringValue(project.Name)
	state.OrganizationID = types.StringValue(project.OrganizationID)
	state.CreationDate = types.StringValue(project.CreationDate.String())
	state.RevisionDate = types.StringValue(project.RevisionDate.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading Project Resource")

	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if p.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	project, err := p.bitwardenClient.Projects().Get(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Project with id: "+state.ID.ValueString(),
			err.Error(),
		)
		return
	}

	state.Name = types.StringValue(project.Name)
	state.OrganizationID = types.StringValue(project.OrganizationID)
	state.CreationDate = types.StringValue(project.CreationDate.String())
	state.RevisionDate = types.StringValue(project.RevisionDate.String())

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state projectResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if p.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	project, err := p.bitwardenClient.Projects().Update(
		state.ID.ValueString(),
		p.organizationId,
		plan.Name.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Project",
			err.Error(),
		)
		return
	}

	state.Name = types.StringValue(project.Name)
	state.OrganizationID = types.StringValue(project.OrganizationID)
	state.CreationDate = types.StringValue(project.CreationDate.String())
	state.RevisionDate = types.StringValue(project.RevisionDate.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if p.bitwardenClient == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	projectDeleteResponse, err := p.bitwardenClient.Projects().Delete([]string{state.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Project",
			err.Error(),
		)
		return
	}
	if projectDeleteResponse.Data[0].Error != nil {
		resp.Diagnostics.AddError(
			"Error deleting Project",
			*projectDeleteResponse.Data[0].Error,
		)
	}
}

func (p *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
