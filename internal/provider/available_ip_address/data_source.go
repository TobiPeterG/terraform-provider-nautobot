package available_ip_address

import (
	"context"
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &AvailableIPAddressDataSource{}
	_ datasource.DataSourceWithConfigure = &AvailableIPAddressDataSource{}
)

type AvailableIPAddressDataSource struct {
	client *shared.APIClient
}

type availableIPAddressDataSourceModel struct {
	PrefixID         types.String `tfsdk:"prefix_id"`
	IPAddressRangeID types.String `tfsdk:"ip_address_range_id"`
	IPVersion        types.Int64  `tfsdk:"ip_version"`
	Address          types.String `tfsdk:"address"`
}

func NewAvailableIPAddressDataSource() datasource.DataSource {
	return &AvailableIPAddressDataSource{}
}

func (d *AvailableIPAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip_address"
}

func (d *AvailableIPAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves the next available IP address from a prefix or non-exclusive IP address range in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"prefix_id": dsschema.StringAttribute{
				Description: "The ID of the prefix from which to retrieve an available IP. Exactly one source must be set.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id"))},
			},
			"ip_address_range_id": dsschema.StringAttribute{
				Description: "The ID of the non-exclusive IP address range from which to retrieve an available IP. Exactly one source must be set.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.ExactlyOneOf(path.MatchRoot("prefix_id"), path.MatchRoot("ip_address_range_id"))},
			},
			"ip_version": dsschema.Int64Attribute{
				Description: "The version of the IP address (4 or 6).",
				Computed:    true,
			},
			"address": dsschema.StringAttribute{
				Description: "The available IP address.",
				Computed:    true,
			},
		},
	}
}

func (d *AvailableIPAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *AvailableIPAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data availableIPAddressDataSourceModel

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

	prefixID, rangeStart, rangeEnd, err := shared.ResolveAvailableIPSource(ctx, c, data.PrefixID.ValueString(), data.IPAddressRangeID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid available IP source",
			err.Error(),
		)
		return
	}

	availableRequest := c.IpamAPI.IpamPrefixesAvailableIpsList(ctx, prefixID)
	if rangeStart != "" {
		availableRequest = availableRequest.RangeStart(rangeStart).RangeEnd(rangeEnd)
	}
	availableIPs, httpResp, err := availableRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to retrieve available IPs",
			shared.HTTPError(err, httpResp),
		)
		return
	}

	if len(availableIPs) == 0 {
		resp.Diagnostics.AddError(
			"No available IPs",
			fmt.Sprintf("No available IP addresses found for prefix %s", prefixID),
		)
		return
	}

	availableIP := availableIPs[0]

	data.IPVersion = types.Int64Value(int64(availableIP.IpVersion))
	data.Address = types.StringValue(availableIP.Address)
	data.PrefixID = types.StringValue(prefixID)

	tflog.Debug(ctx, "read available IP", map[string]any{
		"prefix_id":  prefixID,
		"ip_version": availableIP.IpVersion,
		"address":    availableIP.Address,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
