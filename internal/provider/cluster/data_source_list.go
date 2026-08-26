package cluster

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &ClustersDataSource{}
	_ datasource.DataSourceWithConfigure = &ClustersDataSource{}
)

type ClustersDataSource struct {
	client *shared.APIClient
}

type clustersDataSourceModel struct {
	Clusters []clusterItemModel `tfsdk:"clusters"`
}

func NewClustersDataSource() datasource.DataSource {
	return &ClustersDataSource{}
}

func (d *ClustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *ClustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all clusters in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"clusters": dsschema.ListNestedAttribute{
				Description:  "List of clusters.",
				Computed:     true,
				NestedObject: dsschema.NestedAttributeObject{Attributes: clusterDataAttributes(false)},
			},
		},
	}
}

func (d *ClustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clustersDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	state.Clusters = make([]clusterItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.Cluster], error) {
		rsp, httpResp, err := c.VirtualizationAPI.
			VirtualizationClustersList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.Cluster]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.Cluster]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list clusters", err.Error())
		return
	}

	for _, cluster := range results {
		item, err := clusterModelFromAPI(&cluster)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid cluster data",
				err.Error()+" (name: "+cluster.Name+")",
			)
			return
		}
		state.Clusters = append(state.Clusters, item)
	}

	tflog.Debug(ctx, "read clusters", map[string]any{"count": len(state.Clusters)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
