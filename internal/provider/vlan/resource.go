package vlan

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &VLANResource{}
var _ resource.ResourceWithImportState = &VLANResource{}

type VLANResource struct {
	client *shared.APIClient
}

type vlanModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	VID         types.Int64  `tfsdk:"vid"`
	Description types.String `tfsdk:"description"`
	VLANGroupID types.String `tfsdk:"vlan_group_id"`
	Status      types.String `tfsdk:"status"`
	TenantID    types.String `tfsdk:"tenant_id"`
	RoleID      types.String `tfsdk:"role_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`

	Created types.String `tfsdk:"created"`
}

func NewVLANResource() resource.Resource {
	return &VLANResource{}
}

func (r *VLANResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *VLANResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a VLAN in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "VLAN UUID.",
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "VLAN name.",
			},

			"vid": rschema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4094),
				},
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the VLAN (name).",
			},

			"description": shared.OptionalStringWithDefault("Description of the VLAN."),

			"vlan_group_id": shared.OptionalStringWithDefault("VLAN group UUID."),

			"tenant_id": shared.OptionalStringWithDefault("Tenant UUID associated with the VLAN."),

			"role_id": shared.OptionalStringWithDefault("Role UUID associated with the VLAN."),

			"tags_ids": shared.OptionalStringListWithDefault("Tags associated with the VLAN."),

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VLANResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *VLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	statusID, err := shared.GetStatusID(ctx, c, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status ID", err.Error())
		return
	}

	var vlan nb.VLANRequest
	vlan.Name = plan.Name.ValueString()
	vlan.Vid = int32(plan.VID.ValueInt64())

	vlan.Status = shared.APIReference(statusID)

	if !plan.Description.IsNull() {
		d := plan.Description.ValueString()
		vlan.Description = &d
	}

	if plan.VLANGroupID.ValueString() != "" {
		vlan.VlanGroup = shared.NullableReference(plan.VLANGroupID.ValueString())
	}

	if plan.TenantID.ValueString() != "" {
		vlan.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	if plan.RoleID.ValueString() != "" {
		vlan.Role = shared.NullableReference(plan.RoleID.ValueString())
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if !resp.Diagnostics.HasError() {
			vlan.Tags = shared.APIReferences(tagIDs)
		}
	}

	out, httpResp, err := c.IpamAPI.
		IpamVlansCreate(ctx).
		VLANRequest(vlan).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VLAN", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created VLAN returned no id")
		return
	}

	model, found, diags := r.buildStateModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read VLAN", "created VLAN was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.buildStateModel(ctx, id)
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

func (r *VLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanModel
	var state vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlanID := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedVLANRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	if !plan.VID.Equal(state.VID) {
		v := int32(plan.VID.ValueInt64())
		patch.Vid = &v
	}

	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := shared.GetStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to get status ID", err.Error())
			return
		}
		patch.Status = shared.APIReferencePointer(statusID)
	}

	if !plan.VLANGroupID.Equal(state.VLANGroupID) {
		patch.VlanGroup = shared.NullableReference(plan.VLANGroupID.ValueString())
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = shared.NullableReference(plan.RoleID.ValueString())
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Tags = shared.APIReferences(tagIDs)
	}

	_, httpResp, err := c.IpamAPI.
		IpamVlansPartialUpdate(ctx, vlanID).
		PatchedVLANRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update VLAN", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.buildStateModel(ctx, vlanID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read VLAN", "updated VLAN was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamVlansDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete VLAN", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *VLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VLANResource) buildStateModel(ctx context.Context, id string) (vlanModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, httpResp, err := r.client.Client.IpamAPI.
		IpamVlansRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return vlanModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read VLAN", shared.HTTPError(err, httpResp))
		return vlanModel{}, false, diags
	}
	if v == nil {
		diags.AddError("Invalid API response", "VLAN response is nil")
		return vlanModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("VLAN", id, v.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return vlanModel{}, false, diags
	}

	var m vlanModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(v.Name)
	m.VID = types.Int64Value(int64(v.Vid))

	m.Description = types.StringValue(shared.DerefString(v.Description))

	m.VLANGroupID = shared.NullableReferenceID(v.VlanGroup)

	statusName := ""
	if v.Status.Id != nil && v.Status.Id.String != nil && *v.Status.Id.String != "" {
		n, err := shared.GetStatusName(ctx, r.client.Client, *v.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve VLAN status", err.Error())
			return vlanModel{}, false, diags
		}
		statusName = n
	}
	m.Status = types.StringValue(statusName)

	m.TenantID = shared.NullableReferenceID(v.Tenant)
	m.RoleID = shared.NullableReferenceID(v.Role)

	m.TagsIDs = shared.ReferenceIDs(v.Tags)

	m.Created = shared.NullableTimeValue(v.Created)

	tflog.Debug(ctx, "read VLAN", map[string]any{"id": id})
	return m, true, diags
}
