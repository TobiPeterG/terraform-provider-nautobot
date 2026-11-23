package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	client *APIClient
}

type availableIPModel struct {
	ID          types.String `tfsdk:"id"`
	PrefixID    types.String `tfsdk:"prefix_id"`
	Address     types.String `tfsdk:"address"`
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
				Required:    true,
				Description: "ID of the prefix to allocate the IP address from.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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

			"address": rschema.StringAttribute{
				Computed:    true,
				Description: "Allocated IP address in CIDR notation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"host": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 host address without mask.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"mask_length": rschema.Int64Attribute{
				Computed:    true,
				Description: "Length of the network mask, in bits.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version of the allocated IP address (4 or 6).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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

func (r *AvailableIPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

// randomBackoff returns an increasing backoff with jitter.
func randomBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	const (
		base     = 250 * time.Millisecond
		capDelay = 20 * time.Second
	)

	backoff := base << attempt
	if backoff > capDelay {
		backoff = capDelay
	}

	jitterMax := backoff / 2
	if jitterMax <= 0 {
		return backoff
	}
	jitter := time.Duration(rand.Int63n(int64(jitterMax)))

	return backoff + jitter
}

func (r *AvailableIPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availableIPModel
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

	statusRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
		String: stringPtr(statusID),
	}
	status := nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
		Id: statusRef,
	}

	allocReq := nb.IPAllocationRequest{
		Status: status,
	}

	if v := plan.Type.ValueString(); v != "" {
		ipType, err := ipAddressTypeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("invalid IP address type", err.Error())
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
		allocReq.Role = makeFKUser(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		allocReq.Tenant = makeFKUser(v)
	}
	if v := plan.NatInsideID.ValueString(); v != "" {
		allocReq.NatInside = makeNATInside(v)
	}
	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		tags, diags := expandTagRefs(ctx, plan.TagsIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tags) > 0 {
			allocReq.Tags = tags
		}
	}

	body := []nb.IPAllocationRequest{allocReq}

	// Remove when https://github.com/nautobot/nautobot/issues/8297 is fixed.
	const maxRetries = 10
	var alloc []nb.IPAddress
	var httpResp *http.Response

	for attempt := 0; attempt < maxRetries; attempt++ {
		alloc, httpResp, err = c.IpamAPI.
			IpamPrefixesAvailableIpsCreate(ctx, plan.PrefixID.ValueString()).
			IPAllocationRequest(body).
			Execute()
		if err == nil {
			break
		}

		formattedErr := httpErr(err, httpResp)

		isDuplicate := httpResp != nil &&
			httpResp.StatusCode == http.StatusBadRequest &&
			strings.Contains(formattedErr, "IP address with this Parent and Host already exists")

		if !isDuplicate {
			resp.Diagnostics.AddError("failed to allocate IP address", formattedErr)
			return
		}

		if attempt == maxRetries-1 {
			resp.Diagnostics.AddError("failed to allocate IP address after retries", formattedErr)
			return
		}

		backoff := randomBackoff(attempt)
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("failed to allocate IP address", "context cancelled while retrying after duplicate allocation")
			return
		case <-time.After(backoff):
		}
	}

	if len(alloc) == 0 || alloc[0].Id == nil || *alloc[0].Id == "" {
		resp.Diagnostics.AddError("invalid API response", "no IP address id returned from allocation")
		return
	}

	model, found, diags := r.readModel(ctx, *alloc[0].Id, plan.PrefixID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "created IP address was not found")
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

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString())
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
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}

		statusRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		}
		status := nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: statusRef,
		}
		patch.Status = &status
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

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "updated IP address was not found")
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
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete IP address", httpErr(err, httpResp))
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

func (r *AvailableIPAddressResource) readModel(ctx context.Context, id string, prefixID string) (availableIPModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ip, httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return availableIPModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read IP address", httpErr(err, httpResp))
		return availableIPModel{}, false, diags
	}

	if prefixID == "" {
		prefixID = nullableIPParentStr(ip.Parent).ValueString()
	}
	if prefixID == "" {
		diags.AddError(
			"failed to determine prefix_id for IP address",
			fmt.Sprintf("IP address %s has no parent prefix in Nautobot; import using `<prefix_id>,<ip_id>` if needed", id),
		)
		return availableIPModel{}, false, diags
	}

	statusName := ""
	if ip.Status.Id != nil && ip.Status.Id.String != nil && *ip.Status.Id.String != "" {
		n, err := getStatusName(ctx, r.client.Client, *ip.Status.Id.String)
		if err != nil {
			diags.AddError("failed to read IP address status", err.Error())
			return availableIPModel{}, false, diags
		}
		statusName = n
	}

	var m availableIPModel
	m.ID = types.StringValue(id)
	m.PrefixID = types.StringValue(prefixID)

	m.Address = types.StringValue(ip.Address)
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

	return m, true, diags
}
