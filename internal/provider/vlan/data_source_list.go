package vlan

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &VLANsDataSource{}
	_ datasource.DataSourceWithConfigure = &VLANsDataSource{}
)

type VLANsDataSource struct {
	client *shared.APIClient
}

type vlansDataSourceModel struct {
	VLANs []vlanItemModel `tfsdk:"vlans"`
}

func NewVLANsDataSource() datasource.DataSource {
	return &VLANsDataSource{}
}

func (d *VLANsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlans"
}

func (d *VLANsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all VLANs in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"vlans": dsschema.ListNestedAttribute{
				Description: "List of VLANs.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: vlanDataAttributes(false),
				},
			},
		},
	}
}

func (d *VLANsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VLANsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlansDataSourceModel

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

	state.VLANs = make([]vlanItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.VLAN], error) {
		rsp, httpResp, err := c.IpamAPI.
			IpamVlansList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.VLAN]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.VLAN]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VLANs", err.Error())
		return
	}

	for _, vlan := range results {
		item, err := vlanModelFromAPI(&vlan, statuses.Name)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid VLAN data",
				err.Error()+" (name: "+vlan.Name+")",
			)
			return
		}
		state.VLANs = append(state.VLANs, item)
	}

	tflog.Debug(ctx, "read VLANs", map[string]any{"count": len(state.VLANs)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
