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
	_ datasource.DataSource                     = &VLANGroupDataSource{}
	_ datasource.DataSourceWithConfigure        = &VLANGroupDataSource{}
	_ datasource.DataSourceWithConfigValidators = &VLANGroupDataSource{}
)

var vlanGroupSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type VLANGroupDataSource struct {
	client *shared.APIClient
}

func NewVLANGroupDataSource() datasource.DataSource {
	return &VLANGroupDataSource{}
}

func (d *VLANGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_group"
}

func (d *VLANGroupDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return vlanGroupSelector.ConfigValidators(ctx)
}

func (d *VLANGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific VLAN group in Nautobot by ID or exact name.",
		Attributes:  vlanGroupDataAttributes(true),
	}
}

func (d *VLANGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VLANGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vlanGroupDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	id, name := config.ID.ValueString(), config.Name.ValueString()
	if err := vlanGroupSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid VLAN group selector", err.Error())
		return
	}

	var group *nb.VLANGroup
	if id != "" {
		out, httpResp, err := d.client.Client.IpamAPI.IpamVlanGroupsRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get VLAN group by ID", shared.HTTPError(err, httpResp))
			return
		}
		group = out
	} else {
		out, httpResp, err := d.client.Client.IpamAPI.IpamVlanGroupsList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get VLAN group by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("VLAN group", fmt.Sprintf("name %q", name), len(out.Results)); err != nil {
			resp.Diagnostics.AddError("VLAN group lookup failed", err.Error())
			return
		}
		group = &out.Results[0]
	}
	state, err := vlanGroupDataFromAPI(group)
	if err != nil {
		resp.Diagnostics.AddError("Invalid VLAN group data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("VLAN group", id, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid VLAN group data", err.Error())
		return
	}
	tflog.Debug(ctx, "read VLAN group", map[string]any{"id": state.ID.ValueString(), "name": state.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
