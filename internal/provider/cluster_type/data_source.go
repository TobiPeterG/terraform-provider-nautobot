package cluster_type

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
	_ datasource.DataSource                     = &ClusterTypeDataSource{}
	_ datasource.DataSourceWithConfigure        = &ClusterTypeDataSource{}
	_ datasource.DataSourceWithConfigValidators = &ClusterTypeDataSource{}
)

var clusterTypeSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type ClusterTypeDataSource struct {
	client *shared.APIClient
}

type clusterTypeDataSourceModel = clusterTypeItemModel

func NewClusterTypeDataSource() datasource.DataSource {
	return &ClusterTypeDataSource{}
}

func (d *ClusterTypeDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return clusterTypeSelector.ConfigValidators(ctx)
}

func (d *ClusterTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_type"
}

func (d *ClusterTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific cluster type in Nautobot by ID or exact name.",
		Attributes:  clusterTypeDataAttributes(true),
	}
}

func (d *ClusterTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ClusterTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clusterTypeDataSourceModel

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
	id, name := data.ID.ValueString(), data.Name.ValueString()
	if err := clusterTypeSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid cluster type selector", err.Error())
		return
	}

	var ct *nb.ClusterType
	if id != "" {
		out, httpResp, err := c.VirtualizationAPI.VirtualizationClusterTypesRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get cluster type by ID", shared.HTTPError(err, httpResp))
			return
		}
		ct = out
	} else {
		rsp, httpResp, err := c.VirtualizationAPI.VirtualizationClusterTypesList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get cluster type by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("cluster type", fmt.Sprintf("name %q", name), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Cluster type lookup failed", err.Error())
			return
		}
		ct = &rsp.Results[0]
	}

	data, err := clusterTypeModelFromAPI(ct)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster type data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("cluster type", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid cluster type data", err.Error())
		return
	}

	tflog.Debug(ctx, "read cluster type", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
