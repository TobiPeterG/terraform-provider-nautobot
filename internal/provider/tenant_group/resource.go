package tenant_group

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &TenantGroupResource{}
	_ resource.ResourceWithImportState = &TenantGroupResource{}
)

type TenantGroupResource struct {
	client *shared.APIClient
}

type tenantGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ParentID    types.String `tfsdk:"parent_id"`
	Created     types.String `tfsdk:"created"`
}

func NewTenantGroupResource() resource.Resource {
	return &TenantGroupResource{}
}

func (r *TenantGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_group"
}

func (r *TenantGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a tenant group in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Tenant group's name.",
			},
			"description": shared.OptionalStringWithDefault("Tenant group's description."),
			"parent_id":   shared.OptionalStringWithDefault("UUID of the parent tenant group."),
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TenantGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *TenantGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	body := nb.TenantGroupRequest{
		Name: plan.Name.ValueString(),
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.ParentID.ValueString(); v != "" {
		body.Parent = shared.NullableReference(v)
	}

	out, httpResp, err := c.TenancyAPI.
		TenancyTenantGroupsCreate(ctx).
		TenantGroupRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create tenant group", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created tenant group returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read tenant group", "created tenant group was not found")
		return
	}

	tflog.Debug(ctx, "tenant group created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tenantGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedTenantGroupRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.ParentID.Equal(state.ParentID) {
		patch.Parent = shared.NullableReference(plan.ParentID.ValueString())
	}

	_, httpResp, err := c.TenancyAPI.
		TenancyTenantGroupsPartialUpdate(ctx, id).
		PatchedTenantGroupRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update tenant group", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read tenant group", "updated tenant group was not found")
		return
	}

	tflog.Debug(ctx, "tenant group updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantGroupsDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete tenant group", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *TenantGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TenantGroupResource) readModel(ctx context.Context, id string) (tenantGroupModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	m, httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantGroupsRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return tenantGroupModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read tenant group", shared.HTTPError(err, httpResp))
		return tenantGroupModel{}, false, diags
	}
	if m == nil {
		diags.AddError("Invalid API response", "tenant group response is nil")
		return tenantGroupModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("tenant group", id, m.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return tenantGroupModel{}, false, diags
	}

	var out tenantGroupModel
	out.ID = types.StringValue(id)
	out.Name = types.StringValue(m.Name)

	out.Description = types.StringValue(shared.DerefString(m.Description))
	out.ParentID = shared.NullableReferenceID(m.Parent)
	out.Created = shared.NullableTimeValue(m.Created)

	tflog.Debug(ctx, "read tenant group", map[string]any{"id": id})
	return out, true, diags
}
