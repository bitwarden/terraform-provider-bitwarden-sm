package provider

import (
	"fmt"
	"github.com/bitwarden/sdk-go"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/net/context"
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
	AvoidAmbiguous types.Bool   `tfsdk:"avoid_ambiguous"`
	Length         types.Int64  `tfsdk:"length"`
	Lowercase      types.Bool   `tfsdk:"lowercase"`
	MinLowercase   types.Int64  `tfsdk:"min_lowercase"`
	MinNumber      types.Int64  `tfsdk:"min_number"`
	MinSpecial     types.Int64  `tfsdk:"min_special"`
	MinUppercase   types.Int64  `tfsdk:"min_uppercase"`
	Numbers        types.Bool   `tfsdk:"numbers"`
	Special        types.Bool   `tfsdk:"special"`
	Uppercase      types.Bool   `tfsdk:"uppercase"`
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
			"avoid_ambiguous": schema.BoolAttribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. When set to true, the generated secret will not contain ambiguous characters. The ambiguous characters are: I, O, l, 0, 1.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"length": schema.Int64Attribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. The length of the generated secret. Note that the length of the value must be greater than the sum of all the minimums.",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(64),
				Validators: []validator.Int64{
					int64validator.AtLeastSumOf(path.Expressions{
						path.MatchRoot("min_lowercase"),
						path.MatchRoot("min_uppercase"),
						path.MatchRoot("min_number"),
						path.MatchRoot("min_special"),
					}...),
				},
			},
			"lowercase": schema.BoolAttribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the secret generator to include lowercase characters (a-z).",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"min_lowercase": schema.Int64Attribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the minimum number of lowercase characters in the generated secret. When set, the value must be between 1 and 9. This value is ignored if lowercase is false.",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 9),
				},
			},
			"uppercase": schema.BoolAttribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the secret generator to include uppercase characters (A-Z).",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"min_uppercase": schema.Int64Attribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the minimum number of uppercase characters in the generated secret. When set, the value must be between 1 and 9. This value is ignored if uppercase is false.",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 9),
				},
			},
			"numbers": schema.BoolAttribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the password generator to include numbers (0-9)",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"min_number": schema.Int64Attribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the minimum number of numbers in the generated secret. When set, the value must be between 1 and 9. This value is ignored if numbers is false.",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 9),
				},
			},
			"special": schema.BoolAttribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the password generator to include special characters: ! @ # $ % ^ & *.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"min_special": schema.Int64Attribute{
				Description: "Ignored if value is provided explicitly or secret is updated dynamically in Bitwarden Secrets Manager. Configures the minimum number of special characters in the generated secret. When set, the value must be between 1 and 9. This value is ignored if special is false.",
				Computed:    true,
				Optional:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 9),
				},
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

	var value string
	if plan.Value.IsUnknown() {
		generatedValue, err := createSecretValue(&plan, s.bitwardenClient)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error generating secret value",
				err.Error(),
			)
			return
		}
		value = generatedValue
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
	state.CreationDate = types.StringValue(secret.CreationDate.String())
	state.RevisionDate = types.StringValue(secret.RevisionDate.String())
	state.AvoidAmbiguous = plan.AvoidAmbiguous
	state.Length = plan.Length
	state.Lowercase = plan.Lowercase
	state.MinLowercase = plan.MinLowercase
	state.MinNumber = plan.MinNumber
	state.MinSpecial = plan.MinSpecial
	state.MinUppercase = plan.MinUppercase
	state.Numbers = plan.Numbers
	state.Special = plan.Special
	state.Uppercase = plan.Uppercase

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
	state.CreationDate = types.StringValue(secret.CreationDate.String())
	state.RevisionDate = types.StringValue(secret.RevisionDate.String())

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
		if newGeneratorConfig(&plan, &state) {
			generatedValue, err := createSecretValue(&plan, s.bitwardenClient)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error generating secret value",
					err.Error(),
				)
				return
			}
			value = generatedValue
		} else {
			value = state.Value.ValueString()
		}
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
	state.CreationDate = types.StringValue(secret.CreationDate.String())
	state.RevisionDate = types.StringValue(secret.RevisionDate.String())
	state.AvoidAmbiguous = plan.AvoidAmbiguous
	state.Length = plan.Length
	state.Lowercase = plan.Lowercase
	state.MinLowercase = plan.MinLowercase
	state.MinNumber = plan.MinNumber
	state.MinSpecial = plan.MinSpecial
	state.MinUppercase = plan.MinUppercase
	state.Numbers = plan.Numbers
	state.Special = plan.Special
	state.Uppercase = plan.Uppercase

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

func createSecretValue(config *secretResourceModel, bitwardenClient sdk.BitwardenClientInterface) (string, error) {
	minLowercase := config.MinLowercase.ValueInt64()
	minNumber := config.MinNumber.ValueInt64()
	minSpecial := config.MinSpecial.ValueInt64()
	minUppercase := config.MinUppercase.ValueInt64()

	request := sdk.PasswordGeneratorRequest{
		AvoidAmbiguous: config.AvoidAmbiguous.ValueBool(),
		Length:         config.Length.ValueInt64(),
		Lowercase:      config.Lowercase.ValueBool(),
		MinLowercase:   &minLowercase,
		MinNumber:      &minNumber,
		MinSpecial:     &minSpecial,
		MinUppercase:   &minUppercase,
		Numbers:        config.Numbers.ValueBool(),
		Special:        config.Special.ValueBool(),
		Uppercase:      config.Uppercase.ValueBool(),
	}

	password, err := bitwardenClient.Generators().GeneratePassword(request)
	if err != nil {
		return "", err
	}

	return *password, nil
}

func newGeneratorConfig(plan *secretResourceModel, state *secretResourceModel) bool {
	// Compare all relevant generator configuration attributes between plan and state
	return plan.AvoidAmbiguous.ValueBool() != state.AvoidAmbiguous.ValueBool() ||
		plan.Length.ValueInt64() != state.Length.ValueInt64() ||
		plan.Lowercase.ValueBool() != state.Lowercase.ValueBool() ||
		plan.MinLowercase.ValueInt64() != state.MinLowercase.ValueInt64() ||
		plan.MinNumber.ValueInt64() != state.MinNumber.ValueInt64() ||
		plan.MinSpecial.ValueInt64() != state.MinSpecial.ValueInt64() ||
		plan.MinUppercase.ValueInt64() != state.MinUppercase.ValueInt64() ||
		plan.Numbers.ValueBool() != state.Numbers.ValueBool() ||
		plan.Special.ValueBool() != state.Special.ValueBool() ||
		plan.Uppercase.ValueBool() != state.Uppercase.ValueBool()
}
