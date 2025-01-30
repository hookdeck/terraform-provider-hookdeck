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
	_ resource.Resource                = &typeformSourceResource{}
	_ resource.ResourceWithConfigure   = &typeformSourceResource{}
	_ resource.ResourceWithImportState = &typeformSourceResource{}
)

func init() {
	newResources = append(newResources, NewTypeformSourceResource)
}

func NewTypeformSourceResource() resource.Resource {
	return &typeformSourceResource{}
}

type typeformSourceResource struct {
	client hookdeckClient.Client
}

func (r *typeformSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_typeform_source"
}

func (r *typeformSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Typeform Source Resource",
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

func (r *typeformSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *typeformSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data typeformSourceResourceModel
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

func (r *typeformSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data typeformSourceResourceModel
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

func (r *typeformSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data typeformSourceResourceModel
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

func (r *typeformSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data typeformSourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Source.Delete(context.Background(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting source", err.Error())
	}
}

func (r *typeformSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type typeformSourceResourceModel struct {
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	DisabledAt  types.String `tfsdk:"disabled_at"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	TeamID      types.String `tfsdk:"team_id"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	URL         types.String `tfsdk:"url"`
}

func (m *typeformSourceResourceModel) Refresh(source *hookdeck.Source) {
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

func (m *typeformSourceResourceModel) ToCreatePayload() *hookdeck.SourceCreateRequest {
	return &hookdeck.SourceCreateRequest{
		Description: hookdeck.OptionalOrNull(m.Description.ValueStringPointer()),
		Name:        m.Name.ValueString(),
	}
}

func (m *typeformSourceResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
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
	_ resource.Resource                = &typeformSourceConfigResource{}
	_ resource.ResourceWithConfigure   = &typeformSourceConfigResource{}
	_ resource.ResourceWithImportState = &typeformSourceConfigResource{}
)

func init() {
	newResources = append(newResources, NewTypeformSourceConfigResource)
}

func NewTypeformSourceConfigResource() resource.Resource {
	return &typeformSourceConfigResource{}
}

type typeformSourceConfigResource struct {
	client hookdeckClient.Client
}

func (r *typeformSourceConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_typeform_source_config"
}

func (r *typeformSourceConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Typeform Source Config Resource",
		Attributes: map[string]schema.Attribute{
			"source_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the source",
			},
			
			"auth": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"webhook_secret_key": schema.StringAttribute{
						Required:    true,
						Optional:    false,
						Description: "WebhookSecretKey for Typeform",
						
					},
				},
			},
			
		},
	}
}

func (r *typeformSourceConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *typeformSourceConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data typeformSourceConfigResourceModel
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

func (r *typeformSourceConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data typeformSourceConfigResourceModel
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

func (r *typeformSourceConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data typeformSourceConfigResourceModel
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

func (r *typeformSourceConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data typeformSourceConfigResourceModel
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

func (r *typeformSourceConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type typeformSourceConfigResourceModel struct {
	SourceID types.String           `tfsdk:"source_id"`
	
	Auth     typeformSourceAuthModel `tfsdk:"auth"`
	
}


type typeformSourceAuthModel struct {
	WebhookSecretKey types.String `tfsdk:"webhook_secret_key"`
}


func (m *typeformSourceConfigResourceModel) Refresh(source *hookdeck.Source) {
	m.SourceID = types.StringValue(source.Id)
	
	if source.Verification != nil {
		if source.Verification.Type == "TYPEFORM" {
			data, err := toSourceJSON(source)
			if err != nil {
				// TODO: handle error
				return
			}
			authConfig := getAuthConfig(data)
			
			m.Auth.WebhookSecretKey = types.StringValue(authConfig["webhook_secret_key"].(string))
			
		}
	}
	
}

func (m *typeformSourceConfigResourceModel) ToCreatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("TYPEFORM")
	if err != nil {
		return nil
	}

	

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigTypeform: &hookdeck.SourceTypeConfigTypeform{
			
			Auth: &hookdeck.SourceTypeConfigTypeformAuth{
				
				
				WebhookSecretKey: m.Auth.WebhookSecretKey.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *typeformSourceConfigResourceModel) ToUpdatePayload() *hookdeck.SourceUpdateRequest {
	sourceType, err := hookdeck.NewSourceUpdateRequestTypeFromString("TYPEFORM")
	if err != nil {
		return nil
	}
	

	config := &hookdeck.SourceTypeConfig{
		SourceTypeConfigTypeform: &hookdeck.SourceTypeConfigTypeform{
			
			Auth: &hookdeck.SourceTypeConfigTypeformAuth{
				
				
				WebhookSecretKey: m.Auth.WebhookSecretKey.ValueString(),
				
				
			},
			
		},
	}
	return &hookdeck.SourceUpdateRequest{
		Type:   hookdeck.Optional(sourceType),
		Config: hookdeck.OptionalOrNull(config),
	}
}

func (m *typeformSourceConfigResourceModel) ToDeletePayload() *hookdeck.SourceUpdateRequest {
	return &hookdeck.SourceUpdateRequest{
		Verification: hookdeck.Null[hookdeck.VerificationConfig](),
	}
}

