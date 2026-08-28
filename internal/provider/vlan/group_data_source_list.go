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
	_ datasource.DataSource              = &VLANGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &VLANGroupsDataSource{}
)

type VLANGroupsDataSource struct {
	client *shared.APIClient
}

type vlanGroupsDataSourceModel struct {
	VLANGroups []vlanGroupDataModel `tfsdk:"vlan_groups"`
}

func NewVLANGroupsDataSource() datasource.DataSource {
	return &VLANGroupsDataSource{}
}

func (d *VLANGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_groups"
}

func (d *VLANGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all VLAN groups in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"vlan_groups": dsschema.ListNestedAttribute{
				Description: "List of VLAN groups.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: vlanGroupDataAttributes(false),
				},
			},
		},
	}
}

func (d *VLANGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VLANGroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	state := vlanGroupsDataSourceModel{VLANGroups: make([]vlanGroupDataModel, 0)}
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.VLANGroup], error) {
		out, httpResp, err := d.client.Client.IpamAPI.IpamVlanGroupsList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.VLANGroup]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.VLANGroup]{Items: out.Results, HasNext: out.Next.IsSet() && out.Next.Get() != nil && *out.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VLAN groups", err.Error())
		return
	}
	for i := range results {
		item, err := vlanGroupDataFromAPI(&results[i])
		if err != nil {
			resp.Diagnostics.AddError("Invalid VLAN group data", err.Error())
			return
		}
		state.VLANGroups = append(state.VLANGroups, item)
	}

	tflog.Debug(ctx, "read VLAN groups", map[string]any{"count": len(state.VLANGroups)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
