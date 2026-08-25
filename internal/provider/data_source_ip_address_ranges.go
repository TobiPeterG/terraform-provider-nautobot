package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ datasource.DataSource = &IPAddressRangesDataSource{}
var _ datasource.DataSourceWithConfigure = &IPAddressRangesDataSource{}

type IPAddressRangesDataSource struct{ client *APIClient }

type ipAddressRangeItemModel struct {
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

type ipAddressRangesDataSourceModel struct {
	IPRanges []ipAddressRangeItemModel `tfsdk:"ip_address_ranges"`
}

func NewIPAddressRangesDataSource() datasource.DataSource { return &IPAddressRangesDataSource{} }
func (d *IPAddressRangesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_ranges"
}
func (d *IPAddressRangesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData != nil {
		d.client = req.ProviderData.(*APIClient)
	}
}
func (d *IPAddressRangesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all IP address ranges in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"ip_address_ranges": dsschema.ListNestedAttribute{
				Description: "List of IP address ranges.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id":                dsschema.StringAttribute{Computed: true, Description: "IP address range UUID."},
						"start_address":     dsschema.StringAttribute{Computed: true, Description: "First address in the range (inclusive)."},
						"end_address":       dsschema.StringAttribute{Computed: true, Description: "Last address in the range (inclusive)."},
						"namespace_id":      dsschema.StringAttribute{Computed: true, Description: "Namespace UUID inherited from the parent prefix."},
						"parent_id":         dsschema.StringAttribute{Computed: true, Description: "UUID of the containing parent prefix."},
						"name":              dsschema.StringAttribute{Computed: true, Description: "Human-readable name of the range."},
						"description":       dsschema.StringAttribute{Computed: true, Description: "Description of the range."},
						"count_as_utilized": dsschema.BoolAttribute{Computed: true, Description: "Whether the entire range counts as utilized in its parent prefix."},
						"is_exclusive":      dsschema.BoolAttribute{Computed: true, Description: "Whether individual IP addresses are prohibited inside the range."},
						"status":            dsschema.StringAttribute{Computed: true, Description: "Status of the range (name)."},
						"role_id":           dsschema.StringAttribute{Computed: true, Description: "Role UUID associated with the range."},
						"tenant_id":         dsschema.StringAttribute{Computed: true, Description: "Tenant UUID associated with the range."},
						"tags_ids":          dsschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Tag UUIDs associated with the range."},
						"start_host":        dsschema.StringAttribute{Computed: true, Description: "Normalized first host address."},
						"end_host":          dsschema.StringAttribute{Computed: true, Description: "Normalized last host address."},
						"size":              dsschema.Int64Attribute{Computed: true, Description: "Number of addresses in the range."},
						"ip_version":        dsschema.Int64Attribute{Computed: true, Description: "IP version (4 or 6)."},
						"created":           dsschema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC3339)."},
						"last_updated":      dsschema.StringAttribute{Computed: true, Description: "Last update timestamp (RFC3339)."},
						"display":           dsschema.StringAttribute{Computed: true, Description: "Human-friendly display value."},
						"url":               dsschema.StringAttribute{Computed: true, Description: "API URL of the range."},
						"natural_slug":      dsschema.StringAttribute{Computed: true, Description: "Natural slug of the range."},
						"notes_url":         dsschema.StringAttribute{Computed: true, Description: "Notes API URL of the range."},
					},
				},
			},
		},
	}
}
func (d *IPAddressRangesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ipAddressRangesDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	const pageLimit int32 = 200
	var offset int32 = 0

	state.IPRanges = make([]ipAddressRangeItemModel, 0)

	for {
		rsp, httpResp, err := c.IpamAPI.
			IpamIpAddressRangesList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("start_address").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get IP address ranges list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for i := range results {
			if results[i].Id == nil || *results[i].Id == "" {
				resp.Diagnostics.AddError(
					"Invalid IP address range data",
					"IP address ranges list returned an item with no id.",
				)
				return
			}

			model, diags := d.buildItemModel(ctx, &results[i])
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.IPRanges = append(state.IPRanges, model)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read IP address ranges", map[string]any{"count": len(state.IPRanges)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *IPAddressRangesDataSource) buildItemModel(ctx context.Context, out *nb.IPAddressRange) (ipAddressRangeItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var item ipAddressRangeItemModel

	item.ID = types.StringValue(*out.Id)
	item.StartAddress = types.StringValue(out.StartAddress)
	item.EndAddress = types.StringValue(out.EndAddress)
	item.Name = types.StringValue(derefStr(out.Name))
	item.Description = types.StringValue(derefStr(out.Description))
	item.CountAsUtilized = types.BoolValue(out.CountAsUtilized != nil && *out.CountAsUtilized)
	item.IsExclusive = types.BoolValue(out.IsExclusive != nil && *out.IsExclusive)
	item.RoleID = nullableFKStr(out.Role)
	item.TenantID = nullableFKStr(out.Tenant)
	item.StartHost = types.StringValue(out.StartHost)
	item.EndHost = types.StringValue(out.EndHost)
	item.Size = types.Int64Value(int64(out.Size))
	item.IPVersion = types.Int64Value(int64(out.IpVersion))
	item.Created = nullableTimeStr(out.Created)
	item.LastUpdated = nullableTimeStr(out.LastUpdated)
	item.Display = types.StringValue(out.Display)
	item.URL = types.StringValue(out.Url)
	item.NaturalSlug = types.StringValue(out.NaturalSlug)
	item.NotesURL = types.StringValue(out.NotesUrl)

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		if name, err := getStatusName(ctx, d.client.Client, *out.Status.Id.String); err == nil {
			statusName = name
		}
	}
	item.Status = types.StringValue(statusName)

	tagValues := make([]attr.Value, 0, len(out.Tags))
	for _, tag := range out.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tagValues = append(tagValues, types.StringValue(*tag.Id.String))
		}
	}
	item.TagsIDs = types.ListValueMust(types.StringType, tagValues)

	parentID := ""
	if out.Parent != nil && out.Parent.Id != nil && out.Parent.Id.String != nil {
		parentID = *out.Parent.Id.String
	}
	item.ParentID = types.StringValue(parentID)
	item.NamespaceID = types.StringValue("")
	if parentID != "" {
		parent, httpResp, err := d.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("failed to resolve IP address range namespace", httpErr(err, httpResp))
			return item, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			item.NamespaceID = types.StringValue(*parent.Namespace.Id.String)
		}
	}

	return item, diags
}
