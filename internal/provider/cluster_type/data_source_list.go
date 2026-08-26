package cluster_type

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &ClusterTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &ClusterTypesDataSource{}
)

type ClusterTypesDataSource struct {
	client *shared.APIClient
}

type clusterTypesDataSourceModel struct {
	ClusterTypes []clusterTypeItemModel `tfsdk:"cluster_types"`
}

func NewClusterTypesDataSource() datasource.DataSource {
	return &ClusterTypesDataSource{}
}

func (d *ClusterTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_types"
}

func (d *ClusterTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all cluster types in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"cluster_types": dsschema.ListNestedAttribute{
				Description: "List of cluster types.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: clusterTypeDataAttributes(false),
				},
			},
		},
	}
}

func (d *ClusterTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ClusterTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterTypesDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	state.ClusterTypes = make([]clusterTypeItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.ClusterType], error) {
		rsp, httpResp, err := c.VirtualizationAPI.
			VirtualizationClusterTypesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.ClusterType]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.ClusterType]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cluster types", err.Error())
		return
	}

	for _, ct := range results {
		item, err := clusterTypeModelFromAPI(&ct)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid cluster type data",
				err.Error()+" (name: "+ct.Name+")",
			)
			return
		}
		state.ClusterTypes = append(state.ClusterTypes, item)
	}

	tflog.Debug(ctx, "read cluster types", map[string]any{"count": len(state.ClusterTypes)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
