package cluster

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource                     = &ClusterDataSource{}
	_ datasource.DataSourceWithConfigure        = &ClusterDataSource{}
	_ datasource.DataSourceWithConfigValidators = &ClusterDataSource{}
)

var clusterSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type ClusterDataSource struct {
	client *shared.APIClient
}

type clusterDataSourceModel = clusterItemModel

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

func (d *ClusterDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return clusterSelector.ConfigValidators(ctx)
}

func (d *ClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific cluster in Nautobot by ID or exact name.",
		Attributes:  clusterDataAttributes(true),
	}
}

func (d *ClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clusterDataSourceModel

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

	id, clusterName := data.ID.ValueString(), data.Name.ValueString()
	if err := clusterSelector.Validate(id, map[string]string{"name": clusterName}); err != nil {
		resp.Diagnostics.AddError("Invalid cluster selector", err.Error())
		return
	}

	var cluster *nb.Cluster
	if id != "" {
		var httpResp *http.Response
		var err error
		cluster, httpResp, err = c.VirtualizationAPI.VirtualizationClustersRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get cluster by ID", shared.HTTPError(err, httpResp))
			return
		}
	} else {
		rsp, httpResp, err := c.VirtualizationAPI.VirtualizationClustersList(ctx).Name([]string{clusterName}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get cluster by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("cluster", fmt.Sprintf("name %q", clusterName), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Cluster lookup failed", err.Error())
			return
		}
		cluster = &rsp.Results[0]
	}
	data, err := clusterModelFromAPI(cluster)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cluster data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("cluster", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid cluster data", err.Error())
		return
	}

	tflog.Debug(ctx, "read cluster", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
