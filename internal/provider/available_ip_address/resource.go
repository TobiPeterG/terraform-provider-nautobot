package available_ip_address

import (
	"context"
	"fmt"
	"strings"

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
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &AvailableIPAddressResource{}
	_ resource.ResourceWithImportState = &AvailableIPAddressResource{}
)

type AvailableIPAddressResource struct {
	client *shared.APIClient
}

type availableIPModel struct {
	ID               types.String `tfsdk:"id"`
	PrefixID         types.String `tfsdk:"prefix_id"`
	IPAddressRangeID types.String `tfsdk:"ip_address_range_id"`
	Address          types.String `tfsdk:"address"`
	Host             types.String `tfsdk:"host"`
	MaskLength       types.Int64  `tfsdk:"mask_length"`
	IPVersion        types.Int64  `tfsdk:"ip_version"`
	Type             types.String `tfsdk:"type"`
	DNSName          types.String `tfsdk:"dns_name"`
	Description      types.String `tfsdk:"description"`
	Status           types.String `tfsdk:"status"`
	RoleID           types.String `tfsdk:"role_id"`
	TenantID         types.String `tfsdk:"tenant_id"`
	NatInsideID      types.String `tfsdk:"nat_inside_id"`
	TagsIDs          types.Set    `tfsdk:"tags_ids"`
	Created          types.String `tfsdk:"created"`
}

func NewAvailableIPAddressResource() resource.Resource {
	return &AvailableIPAddressResource{}
}

func (r *AvailableIPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip_address"
}

func (r *AvailableIPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tagsAttr := shared.OptionalStringSetWithDefault("Tags (UUIDs) associated with the IP address.")
	tagsAttr.Validators = []validator.Set{
		setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
	}

	resp.Schema = rschema.Schema{
		Description: "This object allocates and manages an available IP address in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Allocated IP address UUID.",
			},

			"prefix_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the prefix to allocate from. Exactly one of prefix_id or ip_address_range_id must be set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id")),
				},
			},

			"ip_address_range_id": rschema.StringAttribute{
				Optional:      true,
				Description:   "ID of the non-exclusive IP address range to allocate from. Exactly one allocation source must be set.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id")),
				},
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status (by name) of the allocated IP address.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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

			"address": rschema.StringAttribute{
				Computed:    true,
				Description: "Allocated IP address in CIDR notation.",
			},

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
				Description: "IP version of the allocated IP address (4 or 6).",
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC3339).",
			},
		},
	}
}

func (r *AvailableIPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *AvailableIPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availableIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client
	prefixID, rangeStart, rangeEnd, err := shared.ResolveAvailableIPSource(ctx, c, plan.PrefixID.ValueString(), plan.IPAddressRangeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid available IP allocation source", err.Error())
		return
	}

	statusID, err := shared.GetStatusID(ctx, c, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status ID", err.Error())
		return
	}

	allocReq := nb.IPAllocationRequest{
		Status: shared.APIReference(statusID),
	}
	if v := plan.Type.ValueString(); v != "" {
		ipType, err := shared.IPAddressTypeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("Invalid IP address type", err.Error())
			return
		}
		allocReq.Type = ipType
	}
	if v := plan.DNSName.ValueString(); v != "" {
		allocReq.DnsName = &v
	}
	if v := plan.Description.ValueString(); v != "" {
		allocReq.Description = &v
	}
	if v := plan.RoleID.ValueString(); v != "" {
		allocReq.Role = shared.NullableReference(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		allocReq.Tenant = shared.NullableReference(v)
	}
	if v := plan.NatInsideID.ValueString(); v != "" {
		allocReq.NatInside = shared.NullableNATInside(v)
	}
	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		tags, diags := shared.ExpandTagSetRefs(ctx, plan.TagsIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tags) > 0 {
			allocReq.Tags = tags
		}
	}

	allocated, err := r.allocateIPAddress(ctx, prefixID, rangeStart, rangeEnd, allocReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to allocate IP address", err.Error())
		return
	}

	model, found, diags := r.readModel(ctx, *allocated.Id, prefixID, plan.IPAddressRangeID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read IP address", "created IP address was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AvailableIPAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state availableIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString(), state.IPAddressRangeID.ValueString())
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

func (r *AvailableIPAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state availableIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedIPAddressRequest

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

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString(), state.IPAddressRangeID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read IP address", "updated IP address was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AvailableIPAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state availableIPModel
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

func (r *AvailableIPAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rawID := strings.TrimSpace(req.ID)
	if strings.Contains(rawID, ",") {
		parts := strings.SplitN(rawID, ",", 2)
		prefixID := strings.TrimSpace(parts[0])
		ipID := strings.TrimSpace(parts[1])
		if prefixID == "" || ipID == "" {
			resp.Diagnostics.AddError(
				"Invalid import ID",
				"Expected `<ip_id>` or `<prefix_id>,<ip_id>` with non-empty UUIDs.",
			)
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("prefix_id"), prefixID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ipID)...)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AvailableIPAddressResource) readModel(ctx context.Context, id string, prefixID string, rangeID string) (availableIPModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ip, httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return availableIPModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read IP address", shared.HTTPError(err, httpResp))
		return availableIPModel{}, false, diags
	}
	if ip == nil {
		diags.AddError("Invalid API response", "IP address response is nil")
		return availableIPModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("IP address", id, ip.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return availableIPModel{}, false, diags
	}

	if prefixID == "" {
		prefixID = shared.NullableIPParentID(ip.Parent).ValueString()
	}
	if prefixID == "" {
		diags.AddError(
			"Failed to determine prefix_id for IP address",
			fmt.Sprintf("IP address %s has no parent prefix in Nautobot; import using `<prefix_id>,<ip_id>` if needed", id),
		)
		return availableIPModel{}, false, diags
	}

	statusName := ""
	if ip.Status.Id != nil && ip.Status.Id.String != nil && *ip.Status.Id.String != "" {
		n, err := shared.GetStatusName(ctx, r.client.Client, *ip.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve IP address status", err.Error())
			return availableIPModel{}, false, diags
		}
		statusName = n
	}

	var m availableIPModel
	m.ID = types.StringValue(id)
	m.PrefixID = types.StringValue(prefixID)
	if rangeID != "" {
		m.IPAddressRangeID = types.StringValue(rangeID)
	} else {
		m.IPAddressRangeID = types.StringNull()
	}

	m.Address = types.StringValue(ip.Address)
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

	return m, true, diags
}
