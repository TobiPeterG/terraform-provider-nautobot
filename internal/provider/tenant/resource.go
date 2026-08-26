package tenant

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
	_ resource.Resource                = &TenantResource{}
	_ resource.ResourceWithImportState = &TenantResource{}
)

type TenantResource struct {
	client *shared.APIClient
}

type tenantModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Comments      types.String `tfsdk:"comments"`
	TenantGroupID types.String `tfsdk:"tenant_group_id"`
	Created       types.String `tfsdk:"created"`
}

func NewTenantResource() resource.Resource {
	return &TenantResource{}
}

func (r *TenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a tenant in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Tenant's name.",
			},
			"description":     shared.OptionalStringWithDefault("Tenant's description."),
			"comments":        shared.OptionalStringWithDefault("Tenant's comments."),
			"tenant_group_id": shared.OptionalStringWithDefault("UUID of the tenant group this tenant belongs to."),
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *TenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *TenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	body := nb.TenantRequest{
		Name: plan.Name.ValueString(),
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.Comments.ValueString(); v != "" {
		body.Comments = &v
	}
	if v := plan.TenantGroupID.ValueString(); v != "" {
		body.TenantGroup = shared.NullableReference(v)
	}

	out, httpResp, err := c.TenancyAPI.
		TenancyTenantsCreate(ctx).
		TenantRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create tenant", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created tenant returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read tenant", "created tenant was not found")
		return
	}

	tflog.Debug(ctx, "tenant created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
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

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedTenantRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.Comments.Equal(state.Comments) {
		v := plan.Comments.ValueString()
		patch.Comments = &v
	}
	if !plan.TenantGroupID.Equal(state.TenantGroupID) {
		patch.TenantGroup = shared.NullableReference(plan.TenantGroupID.ValueString())
	}

	_, httpResp, err := c.TenancyAPI.
		TenancyTenantsPartialUpdate(ctx, id).
		PatchedTenantRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update tenant", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read tenant", "updated tenant was not found")
		return
	}

	tflog.Debug(ctx, "tenant updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantsDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete tenant", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TenantResource) readModel(ctx context.Context, id string) (tenantModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	m, httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantsRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return tenantModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read tenant", shared.HTTPError(err, httpResp))
		return tenantModel{}, false, diags
	}
	if m == nil {
		diags.AddError("Invalid API response", "tenant response is nil")
		return tenantModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("tenant", id, m.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return tenantModel{}, false, diags
	}

	var out tenantModel
	out.ID = types.StringValue(id)
	out.Name = types.StringValue(m.Name)

	out.Description = types.StringValue(shared.DerefString(m.Description))
	out.Comments = types.StringValue(shared.DerefString(m.Comments))
	out.TenantGroupID = shared.NullableReferenceID(m.TenantGroup)
	out.Created = shared.NullableTimeValue(m.Created)

	tflog.Debug(ctx, "read tenant", map[string]any{"id": id})
	return out, true, diags
}
