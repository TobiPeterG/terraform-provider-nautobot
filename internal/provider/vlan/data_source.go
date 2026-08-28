package vlan

import (
	"context"
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource                     = &VLANDataSource{}
	_ datasource.DataSourceWithConfigure        = &VLANDataSource{}
	_ datasource.DataSourceWithConfigValidators = &VLANDataSource{}
)

var vlanSelector = shared.SelectorSpec{
	NaturalKey: []string{"name"},
	Qualifiers: []string{"vlan_group_id"},
}

type VLANDataSource struct {
	client *shared.APIClient
}

type vlanDataSourceModel = vlanItemModel

func NewVLANDataSource() datasource.DataSource {
	return &VLANDataSource{}
}

func (d *VLANDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return vlanSelector.ConfigValidators(ctx)
}

func (d *VLANDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *VLANDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific VLAN in Nautobot by ID or by exact name and VLAN group.",
		Attributes:  vlanDataAttributes(true),
	}
}

func (d *VLANDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VLANDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vlanDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	id, vlanName, vlanGroupID := data.ID.ValueString(), data.Name.ValueString(), data.VLANGroupID.ValueString()
	if err := vlanSelector.Validate(id, map[string]string{
		"name":          vlanName,
		"vlan_group_id": vlanGroupID,
	}); err != nil {
		resp.Diagnostics.AddError("Invalid VLAN selector", err.Error())
		return
	}

	var vlan *nb.VLAN
	if id != "" {
		out, httpResp, err := c.IpamAPI.IpamVlansRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get VLAN by ID", shared.HTTPError(err, httpResp))
			return
		}
		vlan = out
	} else {
		query := c.IpamAPI.IpamVlansList(ctx).Name([]string{vlanName})
		selectorDescription := fmt.Sprintf("name %q with no VLAN group", vlanName)
		if vlanGroupID != "" {
			query = query.VlanGroup([]string{vlanGroupID})
			selectorDescription = fmt.Sprintf("name %q in VLAN group %q", vlanName, vlanGroupID)
		} else {
			query = query.VlanGroupIsnull(true)
		}
		rsp, httpResp, err := query.Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get VLAN by natural key", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("VLAN", selectorDescription, len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("VLAN lookup failed", err.Error())
			return
		}
		vlan = &rsp.Results[0]
	}

	data, err := vlanModelFromAPI(vlan, func(id string) (string, error) { return shared.GetStatusName(ctx, c, id) })
	if err != nil {
		resp.Diagnostics.AddError("Invalid VLAN data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("VLAN", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid VLAN data", err.Error())
		return
	}

	tflog.Debug(ctx, "read VLAN", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
