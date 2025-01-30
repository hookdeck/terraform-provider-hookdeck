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
	
	
	"github.com/hashicorp/terraform-plugin-framework/types"
	hookdeck "github.com/hookdeck/hookdeck-go-sdk"
	hookdeckClient "github.com/hookdeck/hookdeck-go-sdk/client"
)

// ============================================================================
// Source Resource
// ============================================================================


// Base Source Resource
var (
	_ resource.Resource                = &apikeySourceResource{}
	_ resource.ResourceWithConfigure   = &apikeySourceResource{}
	_ resource.ResourceWithImportState = &apikeySourceResource{}
)

func init() {
	newResources = append(newResources, NewAPIKeySourceResource)
}

func NewAPIKeySourceResource() resource.Resource {
	return &apikeySourceResource{}
}

type apikeySourceResource struct {
	client hookdeckClient.Client
}

func (r *apikeySourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key_source"
}

func (r *apikeySourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "APIKey Source Resource",
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

func (r *apikeySourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apikeySourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data apikeySourceResourceModel
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

func (r *apikeySourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apikeySourceResourceModel
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

func (r *apikeySourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data apikeySourceResourceModel
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

func (r *apikeySourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data apikeySourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Delete(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source", err.Error())
	}
}

func (r *apikeySourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type apikeySourceResourceModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	DisabledAt  types.String `tfsdk:"disabled_at"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	TeamID      types.String `tfsdk:"team_id"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	URL         types.String `tfsdk:"url"`
}

func (m *apikeySourceResourceModel) Refresh(source *hookdeck.Source) {
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

func (m *apikeySourceResourceModel) ToCreatePayload() *hookdeck.SourceCreateRequest {
	return &hookdeck.SourceCreateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        m.Name.ValueString(),
	}
}

func (m *apikeySourceResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
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
	_ resource.Resource                = &apikeySourceConfigResource{}
	_ resource.ResourceWithConfigure   = &apikeySourceConfigResource{}
	_ resource.ResourceWithImportState = &apikeySourceConfigResource{}
)

func init() {
	newResources = append(newResources, NewAPIKeySourceConfigResource)
}

func NewAPIKeySourceConfigResource() resource.Resource {
	return &apikeySourceConfigResource{}
}

type apikeySourceConfigResource struct {
	client hookdeckClient.Client
}

func (r *apikeySourceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key_source_config"
}

func (r *apikeySourceConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "APIKey Source Config Resource",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the source",
			},
			
			"auth": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "ApiKey for APIKey",
						
					},
					"header_key": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "HeaderKey for APIKey",
						
					},
				},
			},
			
		},
	}
}

func (r *apikeySourceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *apikeySourceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data apikeySourceConfigResourceModel
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

func (r *apikeySourceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apikeySourceConfigResourceModel
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

func (r *apikeySourceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data apikeySourceConfigResourceModel
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

func (r *apikeySourceConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data apikeySourceConfigResourceModel
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

func (r *apikeySourceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type apikeySourceConfigResourceModel struct {
	SourceID types.String           `tfsdk:"source_id"`
	
	Auth     apikeySourceAuthModel `tfsdk:"auth"`
	
}


type apikeySourceAuthModel struct {
	ApiKey types.String `tfsdk:"api_key"`
	HeaderKey types.String `tfsdk:"header_key"`
}


func (m *apikeySourceConfigResourceModel) Refresh(source *hookdeck.Source) {
	m.SourceID = types.StringValue(source.Id)
	
	if source.Verification != nil {
		if source.Verification.Type == "APIKEY" {
			data, err := toSourceJSON(source)
			if err != nil {
				// TODO: handle error
				return
			}
			authConfig := getAuthConfig(data)
			
			m.Auth.ApiKey = types.StringValue(authConfig["api_key"].(string))
			
			
			m.Auth.HeaderKey = types.StringValue(authConfig["header_key"].(string))
			
		}
	}
	
}

func (m *apikeySourceConfigResourceModel) ToCreatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("APIKEY")
	if err != nil {
		return nil
	}

	

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigApiKey: &hookdeck.SourceTypeConfigApiKey{
			
			Auth: &hookdeck.SourceTypeConfigApiKeyAuth{
				
				
				ApiKey: m.Auth.ApiKey.ValueString(),
				
				
				
				
				HeaderKey: m.Auth.HeaderKey.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *apikeySourceConfigResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("APIKEY")
	if err != nil {
		return nil
	}
	

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigApiKey: &hookdeck.SourceTypeConfigApiKey{
			
			Auth: &hookdeck.SourceTypeConfigApiKeyAuth{
				
				
				ApiKey: m.Auth.ApiKey.ValueString(),
				
				
				
				
				HeaderKey: m.Auth.HeaderKey.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *apikeySourceConfigResourceModel) ToDeletePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}
}

