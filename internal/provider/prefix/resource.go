package prefix

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

var _ resource.Resource = &PrefixResource{}
var _ resource.ResourceWithImportState = &PrefixResource{}

type PrefixResource struct {
	client *shared.APIClient
}

type prefixModel struct {
	ID            types.String          `tfsdk:"id"`
	VLANID        types.String          `tfsdk:"vlan_id"`
	Prefix        types.String          `tfsdk:"prefix"`
	Description   types.String          `tfsdk:"description"`
	Status        types.String          `tfsdk:"status"`
	ParentID      types.String          `tfsdk:"parent_id"`
	RoleID        types.String          `tfsdk:"role_id"`
	TenantID      types.String          `tfsdk:"tenant_id"`
	RIRID         types.String          `tfsdk:"rir_id"`
	NamespaceID   types.String          `tfsdk:"namespace_id"`
	Created       types.String          `tfsdk:"created"`
	Network       types.String          `tfsdk:"network"`
	Broadcast     types.String          `tfsdk:"broadcast"`
	PrefixLength  types.Int64           `tfsdk:"prefix_length"`
	IPVersion     types.Int64           `tfsdk:"ip_version"`
	DateAllocated shared.RFC3339Instant `tfsdk:"date_allocated"`
	TagsIDs       types.List            `tfsdk:"tags_ids"`
}

func NewPrefixResource() resource.Resource {
	return &PrefixResource{}
}

func (r *PrefixResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefix"
}

func (r *PrefixResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a prefix in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Prefix UUID.",
			},

			"prefix": rschema.StringAttribute{
				Required:    true,
				Description: "The prefix in CIDR notation (e.g. 10.0.0.0/24).",
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the prefix (name).",
			},

			"description": shared.OptionalStringWithDefault("Description of the prefix."),

			"vlan_id": shared.OptionalStringWithDefault("VLAN UUID associated with this prefix."),

			"tenant_id": shared.OptionalStringWithDefault("Tenant UUID associated with the prefix."),

			"role_id": shared.OptionalStringWithDefault("Role UUID associated with the prefix."),

			"parent_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Parent prefix UUID.",
			},

			"rir_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "RIR UUID associated with the prefix.",
			},

			"date_allocated": rschema.StringAttribute{
				Optional:    true,
				CustomType:  shared.RFC3339InstantType{},
				Description: "Date this prefix was allocated/reserved (RFC3339).",
			},

			"namespace_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Namespace UUID associated with the prefix.",
			},

			"tags_ids": shared.OptionalStringListWithDefault("Tags associated with the prefix."),

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"network": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 network address.",
			},

			"broadcast": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 broadcast address.",
			},

			"prefix_length": rschema.Int64Attribute{
				Computed:    true,
				Description: "Length of the prefix, in bits.",
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version (4 or 6).",
			},
		},
	}
}

func (r *PrefixResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *PrefixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan prefixModel
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

	var pr nb.WritablePrefixRequest
	pr.Prefix = plan.Prefix.ValueString()
	pr.Status = shared.APIReference(statusID)

	if !plan.Description.IsNull() {
		v := plan.Description.ValueString()
		pr.Description = &v
	}

	if plan.VLANID.ValueString() != "" {
		pr.Vlan = shared.NullableReference(plan.VLANID.ValueString())
	}

	if plan.TenantID.ValueString() != "" {
		pr.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	if plan.RoleID.ValueString() != "" {
		pr.Role = shared.NullableReference(plan.RoleID.ValueString())
	}
	if !plan.ParentID.IsUnknown() && plan.ParentID.ValueString() != "" {
		pr.Parent = prefixParentReference(plan.ParentID.ValueString())
	}
	if !plan.RIRID.IsUnknown() && plan.RIRID.ValueString() != "" {
		pr.Rir = prefixRIRReference(plan.RIRID.ValueString())
	}
	if !plan.NamespaceID.IsUnknown() && plan.NamespaceID.ValueString() != "" {
		pr.Namespace = shared.APIReferencePointer(plan.NamespaceID.ValueString())
	}
	if !plan.DateAllocated.IsUnknown() && plan.DateAllocated.ValueString() != "" {
		dateAllocated, err := prefixDateAllocated(plan.DateAllocated.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid prefix allocation date", "date_allocated must use RFC3339 format: "+err.Error())
			return
		}
		pr.DateAllocated = dateAllocated
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if !resp.Diagnostics.HasError() {
			pr.Tags = shared.APIReferences(tagIDs)
		}
	}

	out, httpResp, err := c.IpamAPI.
		IpamPrefixesCreate(ctx).
		WritablePrefixRequest(pr).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create prefix", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created prefix returned no id")
		return
	}

	model, found, diags := r.buildStateModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read prefix", "created prefix was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PrefixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state prefixModel
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

func (r *PrefixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan prefixModel
	var state prefixModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedWritablePrefixRequest

	if !plan.Prefix.Equal(state.Prefix) {
		v := plan.Prefix.ValueString()
		patch.Prefix = &v
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

	if !plan.VLANID.Equal(state.VLANID) {
		patch.Vlan = shared.NullableReference(plan.VLANID.ValueString())
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = shared.NullableReference(plan.RoleID.ValueString())
	}

	if !plan.ParentID.IsUnknown() && !plan.ParentID.Equal(state.ParentID) {
		patch.Parent = prefixParentReference(plan.ParentID.ValueString())
	}

	if !plan.RIRID.IsUnknown() && !plan.RIRID.Equal(state.RIRID) {
		patch.Rir = prefixRIRReference(plan.RIRID.ValueString())
	}

	if !plan.NamespaceID.IsUnknown() && plan.NamespaceID.ValueString() != "" && !plan.NamespaceID.Equal(state.NamespaceID) {
		patch.Namespace = shared.APIReferencePointer(plan.NamespaceID.ValueString())
	}

	if !plan.DateAllocated.IsUnknown() && !plan.DateAllocated.Equal(state.DateAllocated) {
		dateAllocated, err := prefixDateAllocated(plan.DateAllocated.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid prefix allocation date", "date_allocated must use RFC3339 format: "+err.Error())
			return
		}
		patch.DateAllocated = dateAllocated
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
		IpamPrefixesPartialUpdate(ctx, id).
		PatchedWritablePrefixRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update prefix", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.buildStateModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read prefix", "updated prefix was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PrefixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state prefixModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamPrefixesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete prefix", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *PrefixResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PrefixResource) buildStateModel(ctx context.Context, id string) (prefixModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, httpResp, err := r.client.Client.IpamAPI.
		IpamPrefixesRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return prefixModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read prefix", shared.HTTPError(err, httpResp))
		return prefixModel{}, false, diags
	}
	if p == nil {
		diags.AddError("Invalid API response", "prefix response is nil")
		return prefixModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("prefix", id, p.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return prefixModel{}, false, diags
	}

	var m prefixModel
	m.ID = types.StringValue(id)
	m.Prefix = types.StringValue(p.Prefix)

	m.Description = types.StringValue(shared.DerefString(p.Description))

	m.Created = shared.NullableTimeValue(p.Created)

	statusName := ""
	if p.Status.Id != nil && p.Status.Id.String != nil && *p.Status.Id.String != "" {
		n, err := shared.GetStatusName(ctx, r.client.Client, *p.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve prefix status", err.Error())
			return prefixModel{}, false, diags
		}
		statusName = n
	}
	m.Status = types.StringValue(statusName)

	parentID := ""
	if p.Parent.IsSet() {
		if parent := p.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
			parentID = *parent.Id.String
		}
	}
	m.ParentID = types.StringValue(parentID)
	m.TenantID = shared.NullableReferenceID(p.Tenant)
	m.RoleID = shared.NullableReferenceID(p.Role)
	rirID := ""
	if p.Rir.IsSet() {
		if rir := p.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
			rirID = *rir.Id.String
		}
	}
	m.RIRID = types.StringValue(rirID)
	namespaceID := ""
	if p.Namespace != nil && p.Namespace.Id != nil && p.Namespace.Id.String != nil {
		namespaceID = *p.Namespace.Id.String
	}
	m.NamespaceID = types.StringValue(namespaceID)
	m.VLANID = shared.NullableReferenceID(p.Vlan)

	m.Network = types.StringValue(p.Network)
	m.Broadcast = types.StringValue(p.Broadcast)
	m.PrefixLength = types.Int64Value(int64(p.PrefixLength))
	m.IPVersion = types.Int64Value(int64(p.IpVersion))

	m.DateAllocated = shared.NullableRFC3339InstantValue(p.DateAllocated)

	m.TagsIDs = shared.ReferenceIDs(p.Tags)

	tflog.Debug(ctx, "read Prefix", map[string]any{"id": id})
	return m, true, diags
}
