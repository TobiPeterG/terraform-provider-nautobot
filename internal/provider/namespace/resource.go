package namespace

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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &NamespaceResource{}
	_ resource.ResourceWithImportState = &NamespaceResource{}
)

type NamespaceResource struct {
	client *shared.APIClient
}

type namespaceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Created     types.String `tfsdk:"created"`
}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

func (r *NamespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Manages an IPAM namespace in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Namespace UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Namespace name.",
			},
			"description": shared.OptionalStringWithDefault("Namespace description."),
			"location_id": shared.OptionalStringWithDefault("UUID of the location associated with the namespace."),
			"tenant_id":   shared.OptionalStringWithDefault("UUID of the tenant associated with the namespace."),
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Namespace creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *NamespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := nb.NamespaceRequest{Name: plan.Name.ValueString()}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.LocationID.ValueString(); v != "" {
		body.Location = shared.NullableReference(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		body.Tenant = shared.NullableReference(v)
	}

	out, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesCreate(ctx).
		NamespaceRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create namespace", shared.HTTPError(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("Invalid API response", "created namespace returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read namespace", "created namespace was not found")
		return
	}

	tflog.Debug(ctx, "namespace created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namespaceModel
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

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	var patch nb.PatchedNamespaceRequest
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.LocationID.Equal(state.LocationID) {
		patch.Location = shared.NullableReference(plan.LocationID.ValueString())
	}
	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = shared.NullableReference(plan.TenantID.ValueString())
	}

	_, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesPartialUpdate(ctx, id).
		PatchedNamespaceRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update namespace", shared.HTTPError(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("Failed to read namespace", "updated namespace was not found")
		return
	}

	tflog.Debug(ctx, "namespace updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !shared.IsNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("Failed to delete namespace", shared.HTTPError(err, httpResp))
	}
}

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NamespaceResource) readModel(ctx context.Context, id string) (namespaceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	n, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesRetrieve(ctx, id).
		Execute()
	if shared.IsNotFoundResponse(httpResp) {
		return namespaceModel{}, false, diags
	}
	if err != nil {
		diags.AddError("Failed to read namespace", shared.HTTPError(err, httpResp))
		return namespaceModel{}, false, diags
	}
	if n == nil {
		diags.AddError("Invalid API response", "namespace response is nil")
		return namespaceModel{}, false, diags
	}
	if err := shared.ValidateAPIObjectID("namespace", id, n.Id); err != nil {
		diags.AddError("Invalid API response", err.Error())
		return namespaceModel{}, false, diags
	}

	out := namespaceModel{
		ID:          types.StringValue(id),
		Name:        types.StringValue(n.Name),
		Description: types.StringValue(shared.DerefString(n.Description)),
		LocationID:  shared.NullableReferenceID(n.Location),
		TenantID:    shared.NullableReferenceID(n.Tenant),
		Created:     shared.NullableTimeValue(n.Created),
	}

	tflog.Debug(ctx, "read namespace", map[string]any{"id": id})
	return out, true, diags
}
