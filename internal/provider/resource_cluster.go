package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &ClusterResource{}
	_ resource.ResourceWithImportState = &ClusterResource{}
)

type ClusterResource struct {
	client *APIClient
}

type clusterModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Comments       types.String `tfsdk:"comments"`
	ClusterTypeID  types.String `tfsdk:"cluster_type_id"`
	ClusterGroupID types.String `tfsdk:"cluster_group_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	LocationID     types.String `tfsdk:"location_id"`
	TagsIDs        types.List   `tfsdk:"tags_ids"`
	Created        types.String `tfsdk:"created"`
}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

func (r *ClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a cluster in Nautobot",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Cluster's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Cluster's name.",
			},
			"cluster_type_id": rschema.StringAttribute{
				Required:    true,
				Description: "ID of the Cluster's type.",
			},

			"comments": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comments or notes about the cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_group_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "ID of the Cluster's group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "ID of the Tenant associated with the cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"location_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "ID of the Location of the cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tags_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "IDs of the Tags associated with the cluster.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the cluster (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	var body nb.ClusterRequest
	body.Name = plan.Name.ValueString()

	clusterTypeRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
		String: stringPtr(plan.ClusterTypeID.ValueString()),
	}
	body.ClusterType = nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
		Id: clusterTypeRef,
	}

	if v := plan.Comments.ValueString(); v != "" {
		body.Comments = &v
	}

	if v := plan.ClusterGroupID.ValueString(); v != "" {
		cgVal := nb.ApprovalWorkflowUser{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(v),
			},
		}
		var cg nb.NullableApprovalWorkflowUser
		cg.Set(&cgVal)
		body.ClusterGroup = cg
	}

	if v := plan.TenantID.ValueString(); v != "" {
		tenantVal := nb.ApprovalWorkflowUser{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(v),
			},
		}
		var tenant nb.NullableApprovalWorkflowUser
		tenant.Set(&tenantVal)
		body.Tenant = tenant
	}

	if v := plan.LocationID.ValueString(); v != "" {
		locVal := nb.ApprovalWorkflowUser{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(v),
			},
		}
		var loc nb.NullableApprovalWorkflowUser
		loc.Set(&locVal)
		body.Location = loc
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tagIDs) > 0 {
			tags := make([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, 0, len(tagIDs))
			for _, t := range tagIDs {
				if t == "" {
					continue
				}
				tags = append(tags, nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
					Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
						String: stringPtr(t),
					},
				})
			}
			body.Tags = tags
		}
	}

	out, httpResp, err := c.VirtualizationAPI.
		VirtualizationClustersCreate(ctx).
		ClusterRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create cluster", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created cluster returned no id")
		return
	}

	model, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, diags := r.readModel(ctx, state.ID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedClusterRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	if !plan.Comments.Equal(state.Comments) {
		if plan.Comments.ValueString() == "" {
			empty := ""
			patch.Comments = &empty
		} else {
			v := plan.Comments.ValueString()
			patch.Comments = &v
		}
	}

	if !plan.ClusterTypeID.Equal(state.ClusterTypeID) {
		v := plan.ClusterTypeID.ValueString()
		clusterTypeRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(v),
		}
		ct := nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: clusterTypeRef,
		}
		patch.ClusterType = &ct
	}

	if !plan.ClusterGroupID.Equal(state.ClusterGroupID) {
		if plan.ClusterGroupID.ValueString() == "" {
			patch.ClusterGroup.Set(nil)
		} else {
			cgVal := nb.ApprovalWorkflowUser{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(plan.ClusterGroupID.ValueString()),
				},
			}
			var cg nb.NullableApprovalWorkflowUser
			cg.Set(&cgVal)
			patch.ClusterGroup = cg
		}
	}

	if !plan.TenantID.Equal(state.TenantID) {
		if plan.TenantID.ValueString() == "" {
			patch.Tenant.Set(nil)
		} else {
			tVal := nb.ApprovalWorkflowUser{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(plan.TenantID.ValueString()),
				},
			}
			var t nb.NullableApprovalWorkflowUser
			t.Set(&tVal)
			patch.Tenant = t
		}
	}

	if !plan.LocationID.Equal(state.LocationID) {
		if plan.LocationID.ValueString() == "" {
			patch.Location.Set(nil)
		} else {
			lVal := nb.ApprovalWorkflowUser{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(plan.LocationID.ValueString()),
				},
			}
			var l nb.NullableApprovalWorkflowUser
			l.Set(&lVal)
			patch.Location = l
		}
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags := make([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, 0, len(tagIDs))
		for _, t := range tagIDs {
			if t == "" {
				continue
			}
			tags = append(tags, nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(t),
				},
			})
		}
		patch.Tags = tags
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationClustersPartialUpdate(ctx, id).
		PatchedClusterRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update cluster", httpErr(err, httpResp))
		return
	}

	model, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClustersDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to delete cluster", httpErr(err, httpResp))
		return
	}
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ClusterResource) readModel(ctx context.Context, id string) (clusterModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	cl, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClustersRetrieve(ctx, id).
		Execute()
	if err != nil {
		diags.AddError("failed to read cluster", httpErr(err, httpResp))
		return clusterModel{}, diags
	}

	var m clusterModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(cl.Name)

	if cl.ClusterType.Id != nil && cl.ClusterType.Id.String != nil {
		m.ClusterTypeID = types.StringValue(*cl.ClusterType.Id.String)
	} else {
		m.ClusterTypeID = types.StringValue("")
	}

	if cl.Comments != nil {
		m.Comments = types.StringValue(*cl.Comments)
	} else {
		m.Comments = types.StringValue("")
	}

	if cl.ClusterGroup.IsSet() {
		if cg := cl.ClusterGroup.Get(); cg != nil && cg.Id != nil && cg.Id.String != nil {
			m.ClusterGroupID = types.StringValue(*cg.Id.String)
		} else {
			m.ClusterGroupID = types.StringValue("")
		}
	} else {
		m.ClusterGroupID = types.StringValue("")
	}

	if cl.Tenant.IsSet() {
		if t := cl.Tenant.Get(); t != nil && t.Id != nil && t.Id.String != nil {
			m.TenantID = types.StringValue(*t.Id.String)
		} else {
			m.TenantID = types.StringValue("")
		}
	} else {
		m.TenantID = types.StringValue("")
	}

	if cl.Location.IsSet() {
		if l := cl.Location.Get(); l != nil && l.Id != nil && l.Id.String != nil {
			m.LocationID = types.StringValue(*l.Id.String)
		} else {
			m.LocationID = types.StringValue("")
		}
	} else {
		m.LocationID = types.StringValue("")
	}

	if len(cl.Tags) > 0 {
		vals := make([]attr.Value, 0, len(cl.Tags))
		for _, t := range cl.Tags {
			if t.Id != nil && t.Id.String != nil {
				vals = append(vals, types.StringValue(*t.Id.String))
			}
		}
		m.TagsIDs = types.ListValueMust(types.StringType, vals)
	} else {
		m.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if cl.Created.IsSet() && cl.Created.Get() != nil {
		m.Created = types.StringValue(cl.Created.Get().Format(time.RFC3339))
	} else {
		m.Created = types.StringNull()
	}

	return m, diags
}
