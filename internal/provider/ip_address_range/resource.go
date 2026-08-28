package ip_address_range

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &IPAddressRangeResource{}
var _ resource.ResourceWithImportState = &IPAddressRangeResource{}

type IPAddressRangeResource struct{ client *shared.APIClient }

type ipAddressRangeModel struct {
	ID              types.String `tfsdk:"id"`
	StartAddress    types.String `tfsdk:"start_address"`
	EndAddress      types.String `tfsdk:"end_address"`
	NamespaceID     types.String `tfsdk:"namespace_id"`
	ParentID        types.String `tfsdk:"parent_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	CountAsUtilized types.Bool   `tfsdk:"count_as_utilized"`
	IsExclusive     types.Bool   `tfsdk:"is_exclusive"`
	Status          types.String `tfsdk:"status"`
	RoleID          types.String `tfsdk:"role_id"`
	TenantID        types.String `tfsdk:"tenant_id"`
	TagsIDs         types.List   `tfsdk:"tags_ids"`
	StartHost       types.String `tfsdk:"start_host"`
	EndHost         types.String `tfsdk:"end_host"`
	Size            types.Int64  `tfsdk:"size"`
	IPVersion       types.Int64  `tfsdk:"ip_version"`
	Created         types.String `tfsdk:"created"`
}

func NewIPAddressRangeResource() resource.Resource {
	return &IPAddressRangeResource{}
}
func (r *IPAddressRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}
func (r *IPAddressRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *IPAddressRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedString := func(description string) rschema.StringAttribute {
		return rschema.StringAttribute{Computed: true, Description: description}
	}
	stableComputedString := func(description string) rschema.StringAttribute {
		attribute := computedString(description)
		attribute.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		return attribute
	}
	resp.Schema = rschema.Schema{
		Description: "Manages an IP address range in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id":            stableComputedString("IP address range UUID."),
			"start_address": rschema.StringAttribute{Required: true, Description: "First host address in the range (inclusive)."},
			"end_address":   rschema.StringAttribute{Required: true, Description: "Last host address in the range (inclusive)."},
			"namespace_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Namespace UUID used to resolve the parent prefix; provide this and/or parent_id.",
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.Expressions{path.MatchRoot("parent_id")}...),
				},
			},
			"parent_id":         rschema.StringAttribute{Optional: true, Computed: true, Description: "UUID of the containing parent prefix; provide this and/or namespace_id."},
			"name":              shared.OptionalStringWithDefault("Human-readable name of the range."),
			"description":       shared.OptionalStringWithDefault("Description of the range."),
			"count_as_utilized": rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Whether the entire range counts as utilized in its parent prefix."},
			"is_exclusive":      rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Whether creation of individual IP addresses inside the range is prohibited."},
			"status":            rschema.StringAttribute{Required: true, Description: "Status of the range (name)."},
			"role_id":           shared.OptionalStringWithDefault("Role UUID associated with the range."),
			"tenant_id":         shared.OptionalStringWithDefault("Tenant UUID associated with the range."),
			"tags_ids":          shared.OptionalStringListWithDefault("Tag UUIDs associated with the range."),
			"start_host":        computedString("Normalized first host address."),
			"end_host":          computedString("Normalized last host address."),
			"size":              rschema.Int64Attribute{Computed: true, Description: "Number of addresses in the range."},
			"ip_version":        rschema.Int64Attribute{Computed: true, Description: "IP version (4 or 6)."},
			"created":           stableComputedString("Creation timestamp (RFC3339)."),
		},
	}
}

func rangeNamespaceRef(id string) *nb.BulkWritableIPAddressRangeRequestNamespace {
	if id == "" {
		return nil
	}
	return &nb.BulkWritableIPAddressRangeRequestNamespace{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)}}
}

func rangeParentRef(id string) *nb.BulkWritableIPAddressRangeRequestParent {
	if id == "" {
		return nil
	}
	return &nb.BulkWritableIPAddressRangeRequestParent{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)}}
}

func (r *IPAddressRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressRangeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	statusID, err := shared.GetStatusID(ctx, r.client.Client, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status ID", err.Error())
		return
	}
	if plan.NamespaceID.ValueString() == "" && plan.ParentID.ValueString() == "" {
		resp.Diagnostics.AddError("Missing namespace or parent prefix", "at least one of `namespace_id` or `parent_id` must be provided")
		return
	}
	body := nb.IPAddressRangeRequest{StartAddress: plan.StartAddress.ValueString(), EndAddress: plan.EndAddress.ValueString(), Status: shared.APIReference(statusID)}
	body.Namespace = rangeNamespaceRef(plan.NamespaceID.ValueString())
	body.Parent = rangeParentRef(plan.ParentID.ValueString())
	name, description := plan.Name.ValueString(), plan.Description.ValueString()
	body.Name, body.Description = &name, &description
	count, exclusive := plan.CountAsUtilized.ValueBool(), plan.IsExclusive.ValueBool()
	body.CountAsUtilized, body.IsExclusive = &count, &exclusive
	if plan.RoleID.ValueString() != "" {
		body.Role = shared.NullableReference(plan.RoleID.ValueString())
	}
	if plan.TenantID.ValueString() != "" {
		body.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}
	var tagIDs []string
	resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body.Tags = shared.APIReferences(tagIDs)
	out, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesCreate(ctx).IPAddressRangeRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create IP address range", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created IP address range returned no id")
		return
	}
	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if !found && !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to read IP address range", "created range was not found")
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (r *IPAddressRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressRangeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, found, diags := r.readModel(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (r *IPAddressRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressRangeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	patch := nb.PatchedIPAddressRangeRequest{}
	if !plan.StartAddress.Equal(state.StartAddress) {
		v := plan.StartAddress.ValueString()
		patch.StartAddress = &v
	}
	if !plan.EndAddress.Equal(state.EndAddress) {
		v := plan.EndAddress.ValueString()
		patch.EndAddress = &v
	}
	if !plan.NamespaceID.IsUnknown() && !plan.NamespaceID.Equal(state.NamespaceID) {
		patch.Namespace = rangeNamespaceRef(plan.NamespaceID.ValueString())
	}
	if !plan.ParentID.IsUnknown() && !plan.ParentID.Equal(state.ParentID) {
		patch.Parent = rangeParentRef(plan.ParentID.ValueString())
	}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.CountAsUtilized.Equal(state.CountAsUtilized) {
		v := plan.CountAsUtilized.ValueBool()
		patch.CountAsUtilized = &v
	}
	if !plan.IsExclusive.Equal(state.IsExclusive) {
		v := plan.IsExclusive.ValueBool()
		patch.IsExclusive = &v
	}
	if !plan.Status.Equal(state.Status) {
		id, err := shared.GetStatusID(ctx, r.client.Client, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to get status ID", err.Error())
			return
		}
		patch.Status = shared.APIReferencePointer(id)
	}
	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = shared.NullableReference(plan.RoleID.ValueString())
	}
	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}
	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var ids []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Tags = shared.APIReferences(ids)
	}
	_, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesPartialUpdate(ctx, state.ID.ValueString()).PatchedIPAddressRangeRequest(patch).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update IP address range", shared.HTTPError(err, httpResp))
		return
	}
	model, found, diags := r.readModel(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if !found && !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("Failed to read IP address range", "updated range was not found")
	}
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (r *IPAddressRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressRangeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesDestroy(ctx, state.ID.ValueString()).Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete IP address range", shared.HTTPError(err, httpResp))
	}
}
func (r *IPAddressRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *IPAddressRangeResource) readModel(ctx context.Context, id string) (ipAddressRangeModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	out, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesRetrieve(ctx, id).Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return ipAddressRangeModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read IP address range", shared.HTTPError(err, httpResp))
		return ipAddressRangeModel{}, false, diags
	}
	if out == nil {
		diags.AddError("Invalid API response", "IP address range response is nil")
		return ipAddressRangeModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("IP address range", id, out.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return ipAddressRangeModel{}, false, diags
	}

	var model ipAddressRangeModel
	model.ID = types.StringValue(id)
	model.StartAddress = types.StringValue(out.StartAddress)
	model.EndAddress = types.StringValue(out.EndAddress)
	model.Name = types.StringValue(shared.DerefString(out.Name))
	model.Description = types.StringValue(shared.DerefString(out.Description))
	model.CountAsUtilized = types.BoolValue(out.CountAsUtilized != nil && *out.CountAsUtilized)
	model.IsExclusive = types.BoolValue(out.IsExclusive != nil && *out.IsExclusive)
	model.RoleID = shared.NullableReferenceID(out.Role)
	model.TenantID = shared.NullableReferenceID(out.Tenant)
	model.StartHost = types.StringValue(out.StartHost)
	model.EndHost = types.StringValue(out.EndHost)
	model.Size = types.Int64Value(int64(out.Size))
	model.IPVersion = types.Int64Value(int64(out.IpVersion))
	model.Created = shared.NullableTimeValue(out.Created)

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		name, err := shared.GetStatusName(ctx, r.client.Client, *out.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve IP address range status", err.Error())
			return ipAddressRangeModel{}, false, diags
		}
		statusName = name
	}
	model.Status = types.StringValue(statusName)

	model.TagsIDs = shared.ReferenceIDs(out.Tags)

	parentID := ""
	if out.Parent != nil && out.Parent.Id != nil && out.Parent.Id.String != nil {
		parentID = *out.Parent.Id.String
	}
	model.ParentID = types.StringValue(parentID)
	model.NamespaceID = types.StringValue("")
	if parentID != "" {
		parent, parentResp, err := r.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("Failed to resolve IP address range namespace", shared.HTTPError(err, parentResp))
			return model, true, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			model.NamespaceID = types.StringValue(*parent.Namespace.Id.String)
		}
	}

	return model, true, diags
}
