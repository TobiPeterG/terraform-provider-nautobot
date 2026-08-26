package ip_address_range

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ datasource.DataSource = &IPAddressRangesDataSource{}
var _ datasource.DataSourceWithConfigure = &IPAddressRangesDataSource{}

type IPAddressRangesDataSource struct{ client *shared.APIClient }

type ipAddressRangesDataSourceModel struct {
	IPAddressRanges []ipAddressRangeItemModel `tfsdk:"ip_address_ranges"`
}

func NewIPAddressRangesDataSource() datasource.DataSource {
	return &IPAddressRangesDataSource{}
}
func (d *IPAddressRangesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_ranges"
}
func (d *IPAddressRangesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}
func (d *IPAddressRangesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all IP address ranges in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"ip_address_ranges": dsschema.ListNestedAttribute{
				Description: "List of IP address ranges.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: ipAddressRangeDataAttributes(false),
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
	statuses, err := shared.NewStatusResolver(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("Failed to load statuses", err.Error())
		return
	}

	state.IPAddressRanges = make([]ipAddressRangeItemModel, 0)
	namespaceByParentID := make(map[string]string)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.IPAddressRange], error) {
		rsp, httpResp, err := c.IpamAPI.
			IpamIpAddressRangesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.IPAddressRange]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.IPAddressRange]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list IP address ranges", err.Error())
		return
	}

	for i := range results {
		model, diags := d.buildItemModel(ctx, statuses, namespaceByParentID, &results[i])
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.IPAddressRanges = append(state.IPAddressRanges, model)
	}

	tflog.Debug(ctx, "read IP address ranges", map[string]any{"count": len(state.IPAddressRanges)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *IPAddressRangesDataSource) buildItemModel(ctx context.Context, statuses *shared.StatusResolver, namespaceByParentID map[string]string, out *nb.IPAddressRange) (ipAddressRangeItemModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	item, err := ipAddressRangeModelFromAPI(out)
	if err != nil {
		diags.AddError("Invalid IP address range data", err.Error())
		return ipAddressRangeItemModel{}, diags
	}

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		name, err := statuses.Name(*out.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve IP address range status", err.Error())
			return ipAddressRangeItemModel{}, diags
		}
		statusName = name
	}
	item.Status = types.StringValue(statusName)

	parentID := item.ParentID.ValueString()
	if parentID != "" {
		if namespaceID, ok := namespaceByParentID[parentID]; ok {
			item.NamespaceID = types.StringValue(namespaceID)
			return item, diags
		}
		parent, httpResp, err := d.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("Failed to resolve IP address range namespace", shared.HTTPError(err, httpResp))
			return item, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			namespaceID := *parent.Namespace.Id.String
			namespaceByParentID[parentID] = namespaceID
			item.NamespaceID = types.StringValue(namespaceID)
		} else {
			namespaceByParentID[parentID] = ""
		}
	}

	return item, diags
}
