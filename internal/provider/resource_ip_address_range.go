package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &IPAddressRangeResource{}
var _ resource.ResourceWithImportState = &IPAddressRangeResource{}

type IPAddressRangeResource struct{ client *APIClient }

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
	Display         types.String `tfsdk:"display"`
	URL             types.String `tfsdk:"url"`
	NaturalSlug     types.String `tfsdk:"natural_slug"`
	NotesURL        types.String `tfsdk:"notes_url"`
}

func NewIPAddressRangeResource() resource.Resource { return &IPAddressRangeResource{} }
func (r *IPAddressRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}
func (r *IPAddressRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData != nil {
		r.client = req.ProviderData.(*APIClient)
	}
}

func (r *IPAddressRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedString := func(description string) rschema.StringAttribute {
		return rschema.StringAttribute{Computed: true, Description: description, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
	}
	optionalString := func(description string) rschema.StringAttribute {
		return rschema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), Description: description, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
	}
	resp.Schema = rschema.Schema{Description: "This object manages an IP address range in Nautobot.", Attributes: map[string]rschema.Attribute{
		"id":                computedString("IP address range UUID."),
		"start_address":     rschema.StringAttribute{Required: true, Description: "First host address in the range (inclusive)."},
		"end_address":       rschema.StringAttribute{Required: true, Description: "Last host address in the range (inclusive)."},
		"namespace_id":      rschema.StringAttribute{Optional: true, Computed: true, Description: "Namespace UUID used to resolve the parent prefix.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"parent_id":         computedString("UUID of the containing parent prefix."),
		"name":              optionalString("Human-readable name of the range."),
		"description":       optionalString("Description of the range."),
		"count_as_utilized": rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Whether the entire range counts as utilized in its parent prefix."},
		"is_exclusive":      rschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Whether creation of individual IP addresses inside the range is prohibited."},
		"status":            rschema.StringAttribute{Required: true, Description: "Status of the range (name)."},
		"role_id":           optionalString("Role UUID associated with the range."),
		"tenant_id":         optionalString("Tenant UUID associated with the range."),
		"tags_ids":          rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})), Description: "Tag UUIDs associated with the range."},
		"start_host":        computedString("Normalized first host address."),
		"end_host":          computedString("Normalized last host address."),
		"size":              rschema.Int64Attribute{Computed: true, Description: "Number of addresses in the range."},
		"ip_version":        rschema.Int64Attribute{Computed: true, Description: "IP version (4 or 6)."},
		"created":           computedString("Creation timestamp (RFC3339)."),
		"display":           computedString("Human-friendly display value."),
		"url":               computedString("API URL of the range."),
		"natural_slug":      computedString("Natural slug of the range."),
		"notes_url":         computedString("Notes API URL of the range."),
	}}
}

func rangeNamespaceRef(id string) *nb.BulkWritableIPAddressRangeRequestNamespace {
	if id == "" {
		return nil
	}
	return &nb.BulkWritableIPAddressRangeRequestNamespace{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: stringPtr(id)}}
}

func ipAddressRangeTags(ids []string) []nb.BulkWritableCableRequestStatus {
	tags := make([]nb.BulkWritableCableRequestStatus, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			tags = append(tags, nb.BulkWritableCableRequestStatus{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: stringPtr(id)}})
		}
	}
	return tags
}

func (r *IPAddressRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressRangeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	statusID, err := getStatusID(ctx, r.client.Client, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get status id", err.Error())
		return
	}
	body := nb.IPAddressRangeRequest{StartAddress: plan.StartAddress.ValueString(), EndAddress: plan.EndAddress.ValueString(), Status: nb.BulkWritableCableRequestStatus{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: stringPtr(statusID)}}}
	body.Namespace = rangeNamespaceRef(plan.NamespaceID.ValueString())
	name, description := plan.Name.ValueString(), plan.Description.ValueString()
	body.Name, body.Description = &name, &description
	count, exclusive := plan.CountAsUtilized.ValueBool(), plan.IsExclusive.ValueBool()
	body.CountAsUtilized, body.IsExclusive = &count, &exclusive
	if plan.RoleID.ValueString() != "" {
		body.Role = makeFKUser(plan.RoleID.ValueString())
	}
	if plan.TenantID.ValueString() != "" {
		body.Tenant = makeFKUser(plan.TenantID.ValueString())
	}
	var tagIDs []string
	resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body.Tags = ipAddressRangeTags(tagIDs)
	out, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesCreate(ctx).IPAddressRangeRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create IP address range", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created IP address range returned no id")
		return
	}
	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if !found && !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("failed to read IP address range", "created range was not found")
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
	if !plan.NamespaceID.Equal(state.NamespaceID) {
		patch.Namespace = rangeNamespaceRef(plan.NamespaceID.ValueString())
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
		id, err := getStatusID(ctx, r.client.Client, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}
		patch.Status = &nb.BulkWritableCableRequestStatus{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: stringPtr(id)}}
	}
	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = makeFKUser(plan.RoleID.ValueString())
	}
	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = makeFKUser(plan.TenantID.ValueString())
	}
	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var ids []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Tags = ipAddressRangeTags(ids)
	}
	_, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesPartialUpdate(ctx, state.ID.ValueString()).PatchedIPAddressRangeRequest(patch).Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update IP address range", httpErr(err, httpResp))
		return
	}
	model, found, diags := r.readModel(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if !found && !resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError("failed to read IP address range", "updated range was not found")
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
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete IP address range", httpErr(err, httpResp))
	}
}
func (r *IPAddressRangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *IPAddressRangeResource) readModel(ctx context.Context, id string) (ipAddressRangeModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	out, httpResp, err := r.client.Client.IpamAPI.IpamIpAddressRangesRetrieve(ctx, id).Execute()
	if isNotFoundResponse(httpResp) {
		return ipAddressRangeModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read IP address range", httpErr(err, httpResp))
		return ipAddressRangeModel{}, false, diags
	}

	var model ipAddressRangeModel
	model.ID = types.StringValue(id)
	model.StartAddress = types.StringValue(out.StartAddress)
	model.EndAddress = types.StringValue(out.EndAddress)
	model.Name = types.StringValue(derefStr(out.Name))
	model.Description = types.StringValue(derefStr(out.Description))
	model.CountAsUtilized = types.BoolValue(out.CountAsUtilized != nil && *out.CountAsUtilized)
	model.IsExclusive = types.BoolValue(out.IsExclusive != nil && *out.IsExclusive)
	model.RoleID = nullableFKStr(out.Role)
	model.TenantID = nullableFKStr(out.Tenant)
	model.StartHost = types.StringValue(out.StartHost)
	model.EndHost = types.StringValue(out.EndHost)
	model.Size = types.Int64Value(int64(out.Size))
	model.IPVersion = types.Int64Value(int64(out.IpVersion))
	model.Created = nullableTimeStr(out.Created)
	model.Display = types.StringValue(out.Display)
	model.URL = types.StringValue(out.Url)
	model.NaturalSlug = types.StringValue(out.NaturalSlug)
	model.NotesURL = types.StringValue(out.NotesUrl)

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		if name, err := getStatusName(ctx, r.client.Client, *out.Status.Id.String); err == nil {
			statusName = name
		}
	}
	model.Status = types.StringValue(statusName)

	tagValues := make([]attr.Value, 0, len(out.Tags))
	for _, tag := range out.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tagValues = append(tagValues, types.StringValue(*tag.Id.String))
		}
	}
	model.TagsIDs = types.ListValueMust(types.StringType, tagValues)

	parentID := ""
	if out.Parent != nil && out.Parent.Id != nil && out.Parent.Id.String != nil {
		parentID = *out.Parent.Id.String
	}
	model.ParentID = types.StringValue(parentID)
	model.NamespaceID = types.StringValue("")
	if parentID != "" {
		parent, parentResp, err := r.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("failed to resolve IP address range namespace", httpErr(err, parentResp))
			return model, true, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			model.NamespaceID = types.StringValue(*parent.Namespace.Id.String)
		}
	}

	return model, true, diags
}
