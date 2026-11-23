package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	client *APIClient
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
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewIPAddressResource() resource.Resource {
	return &IPAddressResource{}
}

func (r *IPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (r *IPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description: "IPv4 or IPv6 address in CIDR notation (e.g. `10.0.0.10/24`).",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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
					useIPStateUnlessRelocated(),
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
					useIPStateUnlessRelocated(),
				},
			},

			"type": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(nb.IPADDRESSTYPECHOICES_HOST)),
				Description: "Type of the IP address (`host`, `dhcp`, or `slaac`).",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(nb.IPADDRESSTYPECHOICES_HOST),
						string(nb.IPADDRESSTYPECHOICES_DHCP),
						string(nb.IPADDRESSTYPECHOICES_SLAAC),
					),
				},
			},

			"dns_name": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "DNS name associated with the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"role_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Role UUID associated with the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant UUID associated with the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"nat_inside_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "UUID of the inside IP address for which this address is the NAT outside IP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tags_ids": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tags (UUIDs) associated with the IP address.",
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
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
				Description: "IP version of the IP address (4 or 6).",
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Human-friendly display value.",
			},

			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "API URL of the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the IP address.",
			},

			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get status id", err.Error())
		return
	}

	body := nb.IPAddressRequest{
		Address: plan.Address.ValueString(),
		Status: nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(statusID),
			},
		},
	}

	if !plan.ParentID.IsNull() && !plan.ParentID.IsUnknown() && plan.ParentID.ValueString() != "" {
		body.Parent = makeIPParent(plan.ParentID.ValueString())
	}
	if !plan.NamespaceID.IsNull() && !plan.NamespaceID.IsUnknown() && plan.NamespaceID.ValueString() != "" {
		body.Namespace = makeIPNamespace(plan.NamespaceID.ValueString())
	}
	if v := plan.Type.ValueString(); v != "" {
		ipType, err := ipAddressTypeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("invalid IP address type", err.Error())
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
		body.Role = makeFKUser(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		body.Tenant = makeFKUser(v)
	}
	if v := plan.NatInsideID.ValueString(); v != "" {
		body.NatInside = makeNATInside(v)
	}
	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		tags, diags := expandTagRefs(ctx, plan.TagsIDs)
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
		resp.Diagnostics.AddError("failed to create IP address", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created IP address returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "created IP address was not found")
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
		patch.Parent = makeIPParent(plan.ParentID.ValueString())
	}

	if !plan.NamespaceID.IsNull() && !plan.NamespaceID.IsUnknown() {
		nsVal := plan.NamespaceID.ValueString()
		if nsVal != "" && (!plan.NamespaceID.Equal(state.NamespaceID) || (addressChanged && plan.ParentID.IsUnknown())) {
			patch.Namespace = makeIPNamespace(nsVal)
		}
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}
		patch.Status = &nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(statusID),
			},
		}
	}

	if !plan.Type.Equal(state.Type) {
		ipType, err := ipAddressTypeFromString(plan.Type.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("invalid IP address type", err.Error())
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
		patch.Role = makeFKUser(plan.RoleID.ValueString())
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if !plan.NatInsideID.Equal(state.NatInsideID) {
		patch.NatInside = makeNATInside(plan.NatInsideID.ValueString())
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		tags, diags := expandTagRefs(ctx, plan.TagsIDs)
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
		resp.Diagnostics.AddError("failed to update IP address", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "updated IP address was not found")
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
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete IP address", httpErr(err, httpResp))
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
	if isNotFoundResponse(httpResp) {
		return ipAddressModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read IP address", httpErr(err, httpResp))
		return ipAddressModel{}, false, diags
	}

	parentID := nullableIPParentStr(ip.Parent).ValueString()
	namespaceID, err := resolveIPAddressNamespaceID(ctx, r.client.Client, ip)
	if err != nil {
		diags.AddError("failed to resolve namespace for IP address", err.Error())
		return ipAddressModel{}, false, diags
	}

	statusName := ""
	if ip.Status.Id != nil && ip.Status.Id.String != nil && *ip.Status.Id.String != "" {
		n, err := getStatusName(ctx, r.client.Client, *ip.Status.Id.String)
		if err != nil {
			diags.AddError("failed to read IP address status", err.Error())
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
	m.Type = ipAddressTypeToString(ip.Type)
	m.DNSName = types.StringValue(derefStr(ip.DnsName))
	m.Description = types.StringValue(derefStr(ip.Description))
	m.Status = types.StringValue(statusName)

	m.RoleID = nullableFKStr(ip.Role)
	m.TenantID = nullableFKStr(ip.Tenant)
	m.NatInsideID = nullableNATInsideStr(ip.NatInside)
	m.TagsIDs = flattenTagIDs(ip.Tags)

	m.Created = nullableTimeStr(ip.Created)
	m.Display = types.StringValue(ip.Display)
	m.URL = types.StringValue(ip.Url)
	m.NaturalSlug = types.StringValue(ip.NaturalSlug)
	m.NotesURL = types.StringValue(ip.NotesUrl)

	tflog.Debug(ctx, "read IP address", map[string]any{"id": id})
	return m, true, diags
}
