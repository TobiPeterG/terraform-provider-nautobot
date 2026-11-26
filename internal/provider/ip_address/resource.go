package ip_address

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &IPAddressResource{}
	_ resource.ResourceWithImportState = &IPAddressResource{}
)

type IPAddressResource struct {
	client *shared.APIClient
}

type ipAddressModel struct {
	ID          types.String `tfsdk:"id"`
	Address     types.String `tfsdk:"address"`
	ParentID    types.String `tfsdk:"parent_id"`
	NamespaceID types.String `tfsdk:"namespace_id"`
	Host        types.String `tfsdk:"host"`
	MaskLength  types.Int64  `tfsdk:"mask_length"`
	IPVersion   types.Int64  `tfsdk:"ip_version"`
	Type        types.String `tfsdk:"type"`
	DNSName     types.String `tfsdk:"dns_name"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	RoleID      types.String `tfsdk:"role_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	NatInsideID types.String `tfsdk:"nat_inside_id"`
	TagsIDs     types.Set    `tfsdk:"tags_ids"`
	Created     types.String `tfsdk:"created"`
}

func NewIPAddressResource() resource.Resource {
	return &IPAddressResource{}
}

func (r *IPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (r *IPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tagsAttr := shared.OptionalStringSetWithDefault("Tags (UUIDs) associated with the IP address.")
	tagsAttr.Validators = []validator.Set{
		setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}

	resp.Schema = rschema.Schema{
		Description: "This object manages an IP address in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "IP address UUID.",
			},

			"address": rschema.StringAttribute{
				Required:    true,
				Description: "IPv4 or IPv6 address in CIDR notation (e.g. `10.0.0.10/24`). Changing the host portion forces replacement.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					shared.RequiresReplaceIfIPHostChanged(),
				},
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status (by name) of the IP address.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},

			"parent_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the parent prefix containing this IP address. Automatically determined by Nautobot if omitted.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					shared.UseIPStateUnlessRelocated(),
				},
			},

			"namespace_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the namespace associated with the IP address. Defaults to the `Global` namespace or the parent prefix's namespace when omitted.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					shared.UseIPStateUnlessRelocated(),
				},
			},

			"type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(nb.IPADDRESSTYPECHOICES_HOST)),
				Description: "Type of the IP address (`host`, `dhcp`, or `slaac`; `slaac` is only valid for IPv6 addresses).",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(nb.IPADDRESSTYPECHOICES_HOST),
						string(nb.IPADDRESSTYPECHOICES_DHCP),
						string(nb.IPADDRESSTYPECHOICES_SLAAC),
					),
				},
			},

			"dns_name":      shared.OptionalStringWithDefault("DNS name associated with the IP address."),
			"description":   shared.OptionalStringWithDefault("Description of the IP address."),
			"role_id":       shared.OptionalStringWithDefault("Role UUID associated with the IP address."),
			"tenant_id":     shared.OptionalStringWithDefault("Tenant UUID associated with the IP address."),
			"nat_inside_id": shared.OptionalStringWithDefault("UUID of the inside IP address for which this address is the NAT outside IP."),
			"tags_ids":      tagsAttr,

			"host": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 host address without mask.",
			},

			"mask_length": rschema.Int64Attribute{
				Computed:    true,
				Description: "Length of the network mask, in bits.",
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version of the IP address (4 or 6).",
			},

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

func (r *IPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *IPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressModel
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

	body := nb.IPAddressRequest{
		Address: plan.Address.ValueString(),
		Status:  shared.APIReference(statusID),
	}

	hasParent := !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() && plan.ParentID.ValueString() != ""
	hasNamespace := !plan.NamespaceID.IsNull() && !plan.NamespaceID.IsUnknown() && plan.NamespaceID.ValueString() != ""
	if hasParent {
		body.Parent = shared.NullableIPParent(plan.ParentID.ValueString())
	}
	if hasNamespace {
		body.Namespace = shared.IPNamespaceReference(plan.NamespaceID.ValueString())
	}
	if !hasParent && !hasNamespace {
		globalNSID, err := shared.GetNamespaceIDByName(ctx, c, "Global")
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve default Global namespace", err.Error())
			return
		}
		body.Namespace = shared.IPNamespaceReference(globalNSID)
	}
	if v := plan.Type.ValueString(); v != "" {
		ipType, err := shared.IPAddressTypeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("Invalid IP address type", err.Error())
			return
		}
		body.Type = ipType
	}
	if v := plan.DNSName.ValueString(); v != "" {
		body.DnsName = &v
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.RoleID.ValueString(); v != "" {
		body.Role = shared.NullableReference(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		body.Tenant = shared.NullableReference(v)
	}
	if v := plan.NatInsideID.ValueString(); v != "" {
		body.NatInside = shared.NullableNATInside(v)
	}
	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		tags, diags := shared.ExpandTagSetRefs(ctx, plan.TagsIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tags) > 0 {
			body.Tags = tags
		}
	}

	out, httpResp, err := c.IpamAPI.
		IpamIpAddressesCreate(ctx).
		IPAddressRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create IP address", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created IP address returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read IP address", "created IP address was not found")
		return
	}
	tflog.Debug(ctx, "IP address created", map[string]any{"id": *out.Id, "address": model.Address.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *IPAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressModel
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

func (r *IPAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedIPAddressRequest

	addressChanged := !plan.Address.Equal(state.Address)
	if addressChanged {
		v := plan.Address.ValueString()
		patch.Address = &v
	}

	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() && !plan.ParentID.Equal(state.ParentID) {
		patch.Parent = shared.NullableIPParent(plan.ParentID.ValueString())
	}

	if !plan.NamespaceID.IsNull() && !plan.NamespaceID.IsUnknown() {
		nsVal := plan.NamespaceID.ValueString()
		if nsVal != "" && (!plan.NamespaceID.Equal(state.NamespaceID) || (addressChanged && plan.ParentID.IsUnknown())) {
			patch.Namespace = shared.IPNamespaceReference(nsVal)
		}
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := shared.GetStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to get status ID", err.Error())
			return
		}
		patch.Status = shared.APIReferencePointer(statusID)
	}

	if !plan.Type.Equal(state.Type) {
		ipType, err := shared.IPAddressTypeFromString(plan.Type.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid IP address type", err.Error())
			return
		}
		patch.Type = ipType
	}

	if !plan.DNSName.Equal(state.DNSName) {
		v := plan.DNSName.ValueString()
		patch.DnsName = &v
	}

	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}

	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = shared.NullableReference(plan.RoleID.ValueString())
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	if !plan.NatInsideID.Equal(state.NatInsideID) {
		patch.NatInside = shared.NullableNATInside(plan.NatInsideID.ValueString())
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		tags, diags := shared.ExpandTagSetRefs(ctx, plan.TagsIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Tags = tags
	}

	_, httpResp, err := c.IpamAPI.
		IpamIpAddressesPartialUpdate(ctx, id).
		PatchedIPAddressRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update IP address", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read IP address", "updated IP address was not found")
		return
	}
	tflog.Debug(ctx, "IP address updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *IPAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete IP address", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *IPAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *IPAddressResource) readModel(ctx context.Context, id string) (ipAddressModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ip, httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return ipAddressModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read IP address", shared.HTTPError(err, httpResp))
		return ipAddressModel{}, false, diags
	}
	if ip == nil {
		diags.AddError("Invalid API response", "IP address response is nil")
		return ipAddressModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("IP address", id, ip.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return ipAddressModel{}, false, diags
	}

	parentID := shared.NullableIPParentID(ip.Parent).ValueString()
	namespaceID, err := shared.ResolveIPAddressNamespaceID(ctx, r.client.Client, ip)
	if err != nil {
		diags.AddError("Failed to resolve namespace for IP address", err.Error())
		return ipAddressModel{}, false, diags
	}

	statusName := ""
	if ip.Status.Id != nil && ip.Status.Id.String != nil && *ip.Status.Id.String != "" {
		n, err := shared.GetStatusName(ctx, r.client.Client, *ip.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve IP address status", err.Error())
			return ipAddressModel{}, false, diags
		}
		statusName = n
	}

	var m ipAddressModel
	m.ID = types.StringValue(id)
	m.Address = types.StringValue(ip.Address)
	m.ParentID = types.StringValue(parentID)
	m.NamespaceID = types.StringValue(namespaceID)
	m.Host = types.StringValue(ip.Host)
	m.MaskLength = types.Int64Value(int64(ip.MaskLength))
	m.IPVersion = types.Int64Value(int64(ip.IpVersion))
	m.Type = shared.IPAddressTypeValue(ip.Type)
	m.DNSName = types.StringValue(shared.DerefString(ip.DnsName))
	m.Description = types.StringValue(shared.DerefString(ip.Description))
	m.Status = types.StringValue(statusName)

	m.RoleID = shared.NullableReferenceID(ip.Role)
	m.TenantID = shared.NullableReferenceID(ip.Tenant)
	m.NatInsideID = shared.NullableNATInsideID(ip.NatInside)
	m.TagsIDs = shared.ReferenceIDSet(ip.Tags)

	m.Created = shared.NullableTimeValue(ip.Created)

	tflog.Debug(ctx, "read IP address", map[string]any{"id": id})
	return m, true, diags
}
