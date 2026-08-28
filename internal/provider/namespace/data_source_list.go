package namespace

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &NamespacesDataSource{}
	_ datasource.DataSourceWithConfigure = &NamespacesDataSource{}
)

type NamespacesDataSource struct {
	client *shared.APIClient
}

type namespacesDataSourceModel struct {
	Namespaces []namespaceItemModel `tfsdk:"namespaces"`
}

func NewNamespacesDataSource() datasource.DataSource {
	return &NamespacesDataSource{}
}

func (d *NamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespaces"
}

func (d *NamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all IPAM namespaces in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"namespaces": dsschema.ListNestedAttribute{
				Description: "List of namespaces.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: namespaceDataAttributes(false),
				},
			},
		},
	}
}

func (d *NamespacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *NamespacesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state namespacesDataSourceModel
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	state.Namespaces = make([]namespaceItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.Namespace], error) {
		rsp, httpResp, err := d.client.Client.IpamAPI.
			IpamNamespacesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.Namespace]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.Namespace]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list namespaces", err.Error())
		return
	}
	for _, n := range results {
		item, err := namespaceModelFromAPI(&n)
		if err != nil {
			resp.Diagnostics.AddError("Invalid namespace data", err.Error()+" (name: "+n.Name+")")
			return
		}
		state.Namespaces = append(state.Namespaces, item)
	}

	tflog.Debug(ctx, "read namespaces", map[string]any{"count": len(state.Namespaces)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
