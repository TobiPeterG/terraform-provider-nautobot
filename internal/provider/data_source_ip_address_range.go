package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ datasource.DataSource = &IPAddressRangeDataSource{}
var _ datasource.DataSourceWithConfigure = &IPAddressRangeDataSource{}

type IPAddressRangeDataSource struct{ client *APIClient }

type ipAddressRangeDataSourceModel struct {
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
	LastUpdated     types.String `tfsdk:"last_updated"`
	Display         types.String `tfsdk:"display"`
	URL             types.String `tfsdk:"url"`
	NaturalSlug     types.String `tfsdk:"natural_slug"`
	NotesURL        types.String `tfsdk:"notes_url"`
}

func NewIPAddressRangeDataSource() datasource.DataSource { return &IPAddressRangeDataSource{} }
func (d *IPAddressRangeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}
func (d *IPAddressRangeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*APIClient)
	}
}

func ipAddressRangeDataSourceAttributes() map[string]dsschema.Attribute {
	computedString := func(description string) dsschema.StringAttribute {
		return dsschema.StringAttribute{Computed: true, Description: description}
	}
	return map[string]dsschema.Attribute{
		"id":            dsschema.StringAttribute{Optional: true, Computed: true, Description: "Range UUID. Provide either `id`, or all of `start_address`, `end_address`, and `namespace_id`."},
		"start_address": dsschema.StringAttribute{Optional: true, Computed: true, Description: "First address in the range (inclusive). Provide together with `end_address` and `namespace_id` when `id` is not used."},
		"end_address":   dsschema.StringAttribute{Optional: true, Computed: true, Description: "Last address in the range (inclusive). Provide together with `start_address` and `namespace_id` when `id` is not used."},
		"namespace_id":  dsschema.StringAttribute{Optional: true, Computed: true, Description: "Namespace UUID. Provide together with `start_address` and `end_address` when `id` is not used."},
		"parent_id":     computedString("UUID of the containing parent prefix."), "name": computedString("Range name."), "description": computedString("Range description."),
		"count_as_utilized": dsschema.BoolAttribute{Computed: true, Description: "Whether the range counts as fully utilized."}, "is_exclusive": dsschema.BoolAttribute{Computed: true, Description: "Whether individual addresses are prohibited."},
		"status": computedString("Range status name."), "role_id": computedString("Associated role UUID."), "tenant_id": computedString("Associated tenant UUID."),
		"tags_ids":   dsschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Associated tag UUIDs."},
		"start_host": computedString("Normalized first host."), "end_host": computedString("Normalized last host."),
		"size": dsschema.Int64Attribute{Computed: true, Description: "Number of addresses."}, "ip_version": dsschema.Int64Attribute{Computed: true, Description: "IP version."},
		"created": computedString("Creation timestamp."), "last_updated": computedString("Last update timestamp."), "display": computedString("Display value."), "url": computedString("API URL."), "natural_slug": computedString("Natural slug."), "notes_url": computedString("Notes API URL."),
	}
}

func (d *IPAddressRangeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{Description: "Retrieves an IP address range by ID or by exact boundaries and namespace.", Attributes: ipAddressRangeDataSourceAttributes()}
}

func (d *IPAddressRangeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ipAddressRangeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, start, end, namespace := data.ID.ValueString(), data.StartAddress.ValueString(), data.EndAddress.ValueString(), data.NamespaceID.ValueString()
	byID, byBounds := id != "", start != "" || end != "" || namespace != ""
	if byID == byBounds {
		resp.Diagnostics.AddError("invalid IP address range selector", "Provide either `id`, or all of `start_address`, `end_address`, and `namespace_id`.")
		return
	}
	var outID string
	if byID {
		outID = id
	} else {
		if start == "" || end == "" || namespace == "" {
			resp.Diagnostics.AddError("incomplete IP address range selector", "`start_address`, `end_address`, and `namespace_id` must be provided together.")
			return
		}
		list, httpResp, err := d.client.Client.IpamAPI.IpamIpAddressRangesList(ctx).StartAddress([]string{start}).EndAddress([]string{end}).Namespace([]string{namespace}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("failed to find IP address range", httpErr(err, httpResp))
			return
		}
		if len(list.Results) != 1 {
			resp.Diagnostics.AddError("IP address range lookup failed", fmt.Sprintf("Expected one exact range match, found %d.", len(list.Results)))
			return
		}
		if list.Results[0].Id == nil {
			resp.Diagnostics.AddError("invalid API response", "matched range returned no id")
			return
		}
		outID = *list.Results[0].Id
	}
	out, httpResp, err := d.client.Client.IpamAPI.IpamIpAddressRangesRetrieve(ctx, outID).Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to retrieve IP address range", httpErr(err, httpResp))
		return
	}
	model, diags := d.buildModel(ctx, out)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	}
}

func (d *IPAddressRangeDataSource) buildModel(ctx context.Context, out *nb.IPAddressRange) (ipAddressRangeDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var model ipAddressRangeDataSourceModel

	if out.Id != nil {
		model.ID = types.StringValue(*out.Id)
	}
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
	model.LastUpdated = nullableTimeStr(out.LastUpdated)
	model.Display = types.StringValue(out.Display)
	model.URL = types.StringValue(out.Url)
	model.NaturalSlug = types.StringValue(out.NaturalSlug)
	model.NotesURL = types.StringValue(out.NotesUrl)

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		if name, err := getStatusName(ctx, d.client.Client, *out.Status.Id.String); err == nil {
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
		parent, httpResp, err := d.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("failed to resolve IP address range namespace", httpErr(err, httpResp))
			return model, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			model.NamespaceID = types.StringValue(*parent.Namespace.Id.String)
		}
	}

	return model, diags
}
