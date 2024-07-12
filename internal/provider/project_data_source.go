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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ProjectsDataSource{}
var _ datasource.DataSourceWithConfigure = &ProjectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource defines the data source implementation.
type ProjectsDataSource struct {
	client *sdk.BitwardenClientInterface
}

// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	Projects []projectDataSourceModel `tfsdk:"projects"`
	ID       types.String             `tfsdk:"id"`
}

type projectDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreationDate   types.String `tfsdk:"creation_date"`
	RevisionDate   types.String `tfsdk:"revision_date"`
}

func (d *ProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *ProjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of projects accessible by the machine account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Computed:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of projects accessible by the machine account.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "String identifier of the project.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "String name of the project.",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "String identifier of the organization the projects belongs to.",
							Computed:    true,
						},
						"creation_date": schema.StringAttribute{
							Description: "String representation of the creation date of the project.",
							Computed:    true,
						},
						"revision_date": schema.StringAttribute{
							Description: "String representation of the revision date of the project.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	tflog.Info(ctx, "Configuring Datasource")
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
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	d.client = client

	tflog.Info(ctx, "Datasource Configured")

}

func (d *ProjectsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectsDataSourceModel

	tflog.Info(ctx, "Reading Datasource")

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Client Not Initialized",
			"The Bitwarden client was not properly initialized.",
		)
		return
	}

	client := *d.client
	projects, err := client.Projects().List("secret")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List Projects",
			err.Error(),
		)
		return
	}

	for _, project := range projects.Data {
		projectState := projectDataSourceModel{
			ID:             types.StringValue(project.ID),
			Name:           types.StringValue(project.Name),
			OrganizationID: types.StringValue(project.OrganizationID),
			CreationDate:   types.StringValue(project.CreationDate),
			RevisionDate:   types.StringValue(project.RevisionDate),
		}

		state.Projects = append(state.Projects, projectState)
	}

	state.ID = types.StringValue("placeholder")

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
