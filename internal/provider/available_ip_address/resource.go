package available_ip_address

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	IPVersion        types.Int64  `tfsdk:"ip_version"`
	DNSName          types.String `tfsdk:"dns_name"`
	Status           types.String `tfsdk:"status"`
}

func NewAvailableIPAddressResource() resource.Resource {
	return &AvailableIPAddressResource{}
}

func (r *AvailableIPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip_address"
}

func (r *AvailableIPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id"))},
			},
			"ip_address_range_id": rschema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "ID of the non-exclusive IP address range to allocate from. Exactly one allocation source must be set.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id"))},
			},

			"address": rschema.StringAttribute{
				Computed:    true,
				Description: "Allocated IP address in CIDR notation.",
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version of the allocated IP address (4 or 6).",
			},

			"dns_name": shared.OptionalStringWithDefault("DNS name associated with the IP address."),

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status (by name) of the allocated IP address.",
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

	allocated, err := r.allocateIPAddress(
		ctx, prefixID, rangeStart, rangeEnd, statusID, plan.DNSName.ValueString(),
	)
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

	if !plan.DNSName.Equal(state.DNSName) {
		if plan.DNSName.ValueString() == "" {
			empty := ""
			patch.DnsName = &empty
		} else {
			v := plan.DNSName.ValueString()
			patch.DnsName = &v
		}
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

	// Try to derive prefixID from parent if not provided
	if prefixID == "" && ip.Parent.IsSet() {
		parent := ip.Parent.Get()
		if parent != nil && parent.Id != nil && parent.Id.String != nil && *parent.Id.String != "" {
			prefixID = *parent.Id.String
		}
	}

	var m availableIPModel
	m.ID = types.StringValue(id)
	m.PrefixID = types.StringValue(prefixID)
	m.IPAddressRangeID = types.StringValue(rangeID)

	m.Address = types.StringValue(ip.Address)
	m.IPVersion = types.Int64Value(int64(ip.IpVersion))

	if ip.DnsName != nil {
		m.DNSName = types.StringValue(*ip.DnsName)
	} else {
		m.DNSName = types.StringValue("")
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
	m.Status = types.StringValue(statusName)

	return m, true, diags
}
