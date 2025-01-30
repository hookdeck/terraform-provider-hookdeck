package generated

import (
	"context"
	"fmt"
	"terraform-provider-hookdeck/internal/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// ============================================================================
// Source Resource
// ============================================================================


// Base Source Resource
var (
	_ resource.Resource                = &ebaySourceResource{}
	_ resource.ResourceWithConfigure   = &ebaySourceResource{}
	_ resource.ResourceWithImportState = &ebaySourceResource{}
)

func init() {
	newResources = append(newResources, NewEbaySourceResource)
}

func NewEbaySourceResource() resource.Resource {
	return &ebaySourceResource{}
}

type ebaySourceResource struct {
	client hookdeckClient.Client
}

func (r *ebaySourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ebay_source"
}

func (r *ebaySourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ebay Source Resource",
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Date the source was created",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Description for the source",
			},
			"disabled_at": schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: "Date the source was disabled",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ID of the source",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A unique, human-friendly name for the source",
			},
			"team_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "ID of the workspace",
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					validators.IsRFC3339(),
				},
				Description: "Date the source was last updated",
			},
			"url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "A unique URL that must be supplied to your webhook's provider",
			},
		},
	}
}

func (r *ebaySourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hookdeckClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hookdeckClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = *client
}

func (r *ebaySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ebaySourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.Source.Create(context.Background(), data.ToCreatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error creating source", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ebaySourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.Source.Retrieve(context.Background(), data.ID.ValueString(), &hookdeck.SourceRetrieveRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Error reading source", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ebaySourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.Source.Update(context.Background(), data.ID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error updating source", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ebaySourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Delete(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source", err.Error())
	}
}

func (r *ebaySourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type ebaySourceResourceModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	DisabledAt  types.String `tfsdk:"disabled_at"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	TeamID      types.String `tfsdk:"team_id"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	URL         types.String `tfsdk:"url"`
}

func (m *ebaySourceResourceModel) Refresh(source *hookdeck.Source) {
	m.CreatedAt = types.StringValue(source.CreatedAt.Format(time.RFC3339))
	if source.DisabledAt != nil {
		m.DisabledAt = types.StringValue(source.DisabledAt.Format(time.RFC3339))
	} else {
		m.DisabledAt = types.StringNull()
	}
	m.ID = types.StringValue(source.Id)
	m.Name = types.StringValue(source.Name)
	m.TeamID = types.StringValue(source.TeamId)
	m.UpdatedAt = types.StringValue(source.UpdatedAt.Format(time.RFC3339))
	m.URL = types.StringValue(source.Url)
}

func (m *ebaySourceResourceModel) ToCreatePayload() *hookdeck.SourceCreateRequest {
	return &hookdeck.SourceCreateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        m.Name.ValueString(),
	}
}

func (m *ebaySourceResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        hookdeck.OptionalOrNull(m.Name.ValueStringPointer()),
	}
}


// ============================================================================
// Source Config Resource
// ============================================================================


// Source Config Resource
var (
	_ resource.Resource                = &ebaySourceConfigResource{}
	_ resource.ResourceWithConfigure   = &ebaySourceConfigResource{}
	_ resource.ResourceWithImportState = &ebaySourceConfigResource{}
)

func init() {
	newResources = append(newResources, NewEbaySourceConfigResource)
}

func NewEbaySourceConfigResource() resource.Resource {
	return &ebaySourceConfigResource{}
}

type ebaySourceConfigResource struct {
	client hookdeckClient.Client
}

func (r *ebaySourceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ebay_source_config"
}

func (r *ebaySourceConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ebay Source Config Resource",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the source",
			},
			
			"auth": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"client_secret": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "ClientSecret for Ebay",
						
					},
					"verification_token": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "VerificationToken for Ebay",
						
					},
					"environment": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "Environment for Ebay",
						
						Validators: []validator.String{
							stringvalidator.OneOf(
								"PRODUCTION",
								"SANDBOX",
							),
						},
						
					},
					"dev_id": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "DevId for Ebay",
						
					},
					"client_id": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "ClientId for Ebay",
						
					},
				},
			},
			
		},
	}
}

func (r *ebaySourceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hookdeckClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hookdeckClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = *client
}

func (r *ebaySourceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ebaySourceConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Update(context.Background(), data.SourceID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error creating source config", err.Error())
		return
	}

	source, err := r.client.Source.Retrieve(context.Background(), data.SourceID.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading source config", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ebaySourceConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, err := r.client.Source.Retrieve(context.Background(), data.SourceID.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading source config", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ebaySourceConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Update(context.Background(), data.SourceID.ValueString(), data.ToUpdatePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error updating source config", err.Error())
		return
	}

	source, err := r.client.Source.Retrieve(context.Background(), data.SourceID.ValueString(), &hookdeck.SourceRetrieveRequest{
		Include: hookdeck.String("verification.configs"),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading source config", err.Error())
		return
	}

	data.Refresh(source)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ebaySourceConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ebaySourceConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Update(context.Background(), data.SourceID.ValueString(), data.ToDeletePayload())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source config", err.Error())
		return
	}
}

func (r *ebaySourceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type ebaySourceConfigResourceModel struct {
	SourceID types.String           `tfsdk:"source_id"`
	
	Auth     ebaySourceAuthModel `tfsdk:"auth"`
	
}


type ebaySourceAuthModel struct {
	ClientSecret types.String `tfsdk:"client_secret"`
	VerificationToken types.String `tfsdk:"verification_token"`
	Environment types.String `tfsdk:"environment"`
	DevId types.String `tfsdk:"dev_id"`
	ClientId types.String `tfsdk:"client_id"`
}


func (m *ebaySourceConfigResourceModel) Refresh(source *hookdeck.Source) {
	m.SourceID = types.StringValue(source.Id)
	
	if source.Verification != nil {
		if source.Verification.Type == "EBAY" {
			data, err := toSourceJSON(source)
			if err != nil {
				// TODO: handle error
				return
			}
			authConfig := getAuthConfig(data)
			
			m.Auth.ClientSecret = types.StringValue(authConfig["client_secret"].(string))
			
			
			m.Auth.VerificationToken = types.StringValue(authConfig["verification_token"].(string))
			
			
			m.Auth.Environment = types.StringValue(authConfig["environment"].(string))
			
			
			m.Auth.DevId = types.StringValue(authConfig["dev_id"].(string))
			
			
			m.Auth.ClientId = types.StringValue(authConfig["client_id"].(string))
			
		}
	}
	
}

func (m *ebaySourceConfigResourceModel) ToCreatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("EBAY")
	if err != nil {
		return nil
	}

	
	environment, err := hookdeck.NewSourceTypeConfigEbayAuthEnvironmentFromString(m.Auth.Environment.ValueString())
	if err != nil {
		// TODO: handle error
		return nil
	}

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigEbay: &hookdeck.SourceTypeConfigEbay{
			
			Auth: &hookdeck.SourceTypeConfigEbayAuth{
				
				
				ClientSecret: m.Auth.ClientSecret.ValueString(),
				
				
				
				
				VerificationToken: m.Auth.VerificationToken.ValueString(),
				
				
				
				
				Environment: environment,
				
				
				
				
				DevId: m.Auth.DevId.ValueString(),
				
				
				
				
				ClientId: m.Auth.ClientId.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *ebaySourceConfigResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("EBAY")
	if err != nil {
		return nil
	}
	
	environment, err := hookdeck.NewSourceTypeConfigEbayAuthEnvironmentFromString(m.Auth.Environment.ValueString())
	if err != nil {
		// TODO: handle error
		return nil
	}

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigEbay: &hookdeck.SourceTypeConfigEbay{
			
			Auth: &hookdeck.SourceTypeConfigEbayAuth{
				
				
				ClientSecret: m.Auth.ClientSecret.ValueString(),
				
				
				
				
				VerificationToken: m.Auth.VerificationToken.ValueString(),
				
				
				
				
				Environment: environment,
				
				
				
				
				DevId: m.Auth.DevId.ValueString(),
				
				
				
				
				ClientId: m.Auth.ClientId.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *ebaySourceConfigResourceModel) ToDeletePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}
}

