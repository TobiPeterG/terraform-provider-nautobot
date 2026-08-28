package cluster_type

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
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &ClusterTypeResource{}
	_ resource.ResourceWithImportState = &ClusterTypeResource{}
)

type ClusterTypeResource struct {
	client *shared.APIClient
}

type clusterTypeModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
}

func NewClusterTypeResource() resource.Resource {
	return &ClusterTypeResource{}
}

func (r *ClusterTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_type"
}

func (r *ClusterTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages a cluster type in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Cluster type's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Cluster type's name.",
			},

			"description": shared.OptionalStringWithDefault("Description for the cluster type."),

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the cluster type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *ClusterTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	var body nb.ClusterTypeRequest
	body.Name = plan.Name.ValueString()
	if plan.Description.ValueString() != "" {
		desc := plan.Description.ValueString()
		body.Description = &desc
	}

	out, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesCreate(ctx).
		ClusterTypeRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cluster type", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created cluster type returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read cluster type", "created cluster type was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterTypeModel
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

func (r *ClusterTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedClusterTypeRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.ValueString() == "" {
			empty := ""
			patch.Description = &empty
		} else {
			v := plan.Description.ValueString()
			patch.Description = &v
		}
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesPartialUpdate(ctx, id).
		PatchedClusterTypeRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cluster type", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read cluster type", "updated cluster type was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClusterTypesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete cluster type", shared.HTTPError(err, httpResp))
		return
	}
}

func (r *ClusterTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ClusterTypeResource) readModel(ctx context.Context, id string) (clusterTypeModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ct, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClusterTypesRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return clusterTypeModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read cluster type", shared.HTTPError(err, httpResp))
		return clusterTypeModel{}, false, diags
	}
	if ct == nil {
		diags.AddError("Invalid API response", "cluster type response is nil")
		return clusterTypeModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("cluster type", id, ct.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return clusterTypeModel{}, false, diags
	}

	var m clusterTypeModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(ct.Name)

	m.Description = types.StringValue(shared.DerefString(ct.Description))

	m.Created = shared.NullableTimeValue(ct.Created)

	return m, true, diags
}
