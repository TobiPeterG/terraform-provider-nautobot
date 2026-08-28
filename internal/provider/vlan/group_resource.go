package vlan

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

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
	_ resource.Resource                = &VLANGroupResource{}
	_ resource.ResourceWithImportState = &VLANGroupResource{}
)

type VLANGroupResource struct {
	client *shared.APIClient
}

type vlanGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Range       types.String `tfsdk:"range"`
	LocationID  types.String `tfsdk:"location_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`
	Created     types.String `tfsdk:"created"`
}

func NewVLANGroupResource() resource.Resource {
	return &VLANGroupResource{}
}

func (r *VLANGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_group"
}

func (r *VLANGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a VLAN group in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Description:   "VLAN group UUID.",
			},
			"name":        rschema.StringAttribute{Required: true, Description: "VLAN group name."},
			"description": shared.OptionalStringWithDefault("VLAN group description."),
			"range": rschema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("1-4094"),
				Validators:  []validator.String{vlanRangeValidator{}},
				Description: "Permitted VLAN IDs as a comma-separated list of individual IDs or ranges. Defaults to `1-4094`.",
			},
			"location_id": shared.OptionalStringWithDefault("UUID of the location associated with the VLAN group."),
			"tags_ids":    shared.OptionalStringListWithDefault("UUIDs of tags associated with the VLAN group."),
			"created": rschema.StringAttribute{
				Computed: true, Description: "VLAN group creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *VLANGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *VLANGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := nb.VLANGroupRequest{Name: plan.Name.ValueString()}
	description, vlanRange := plan.Description.ValueString(), plan.Range.ValueString()
	body.Description = &description
	body.Range = &vlanRange
	if locationID := plan.LocationID.ValueString(); locationID != "" {
		body.Location = shared.NullableReference(locationID)
	}
	var tagIDs []string
	resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body.Tags = shared.APIReferences(tagIDs)

	created, httpResp, err := r.client.Client.IpamAPI.IpamVlanGroupsCreate(ctx).VLANGroupRequest(body).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VLAN group", shared.HTTPError(err, httpResp))
		return
	}
	if created.Id == nil || *created.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created VLAN group returned no id")
		return
	}
	model, found, diags := r.readModel(ctx, *created.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read VLAN group", "created VLAN group was not found")
		return
	}
	tflog.Debug(ctx, "VLAN group created", map[string]any{"id": *created.Id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model, found, diags := r.readModel(ctx, state.ID.ValueString())
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

func (r *VLANGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vlanGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var patch nb.PatchedVLANGroupRequest
	if !plan.Name.Equal(state.Name) {
		value := plan.Name.ValueString()
		patch.Name = &value
	}
	if !plan.Description.Equal(state.Description) {
		value := plan.Description.ValueString()
		patch.Description = &value
	}
	if !plan.Range.Equal(state.Range) {
		value := plan.Range.ValueString()
		patch.Range = &value
	}
	if !plan.LocationID.Equal(state.LocationID) {
		patch.Location = shared.NullableReference(plan.LocationID.ValueString())
	}
	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		patch.Tags = shared.APIReferences(tagIDs)
	}

	_, httpResp, err := r.client.Client.IpamAPI.IpamVlanGroupsPartialUpdate(ctx, state.ID.ValueString()).PatchedVLANGroupRequest(patch).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update VLAN group", shared.HTTPError(err, httpResp))
		return
	}
	model, found, diags := r.readModel(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read VLAN group", "updated VLAN group was not found")
		return
	}
	tflog.Debug(ctx, "VLAN group updated", map[string]any{"id": state.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	httpResp, err := r.client.Client.IpamAPI.IpamVlanGroupsDestroy(ctx, state.ID.ValueString()).Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete VLAN group", shared.HTTPError(err, httpResp))
	}
}

func (r *VLANGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VLANGroupResource) readModel(ctx context.Context, id string) (vlanGroupResourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	group, httpResp, err := r.client.Client.IpamAPI.IpamVlanGroupsRetrieve(ctx, id).Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return vlanGroupResourceModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read VLAN group", shared.HTTPError(err, httpResp))
		return vlanGroupResourceModel{}, false, diags
	}
	if group == nil {
		diags.AddError("Invalid API response", "VLAN group response is nil")
		return vlanGroupResourceModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("VLAN group", id, group.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return vlanGroupResourceModel{}, false, diags
	}
	data, err := vlanGroupDataFromAPI(group)
	if err != nil {
		diags.AddError("Invalid API response", err.Error())
		return vlanGroupResourceModel{}, false, diags
	}
	return vlanGroupResourceModel{
		ID: data.ID, Name: data.Name, Description: data.Description, Range: data.Range,
		LocationID: data.LocationID, TagsIDs: data.TagsIDs, Created: data.Created,
	}, true, diags
}
