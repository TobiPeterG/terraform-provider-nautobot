package prefix

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &PrefixesDataSource{}
	_ datasource.DataSourceWithConfigure = &PrefixesDataSource{}
)

type PrefixesDataSource struct {
	client *shared.APIClient
}

type prefixesDataSourceModel struct {
	Prefixes []prefixItemModel `tfsdk:"prefixes"`
}

func NewPrefixesDataSource() datasource.DataSource {
	return &PrefixesDataSource{}
}

func (d *PrefixesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefixes"
}

func (d *PrefixesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all prefixes in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"prefixes": dsschema.ListNestedAttribute{
				Description: "List of prefixes.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: prefixDataAttributes(false),
				},
			},
		},
	}
}

func (d *PrefixesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *PrefixesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state prefixesDataSourceModel

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

	state.Prefixes = make([]prefixItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.Prefix], error) {
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.Prefix]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.Prefix]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list prefixes", err.Error())
		return
	}

	for i := range results {
		prefix := &results[i]
		if prefix.Id == nil || *prefix.Id == "" {
			resp.Diagnostics.AddError(
				"Invalid prefix data",
				"Prefixes list returned an item with no id (prefix: "+prefix.Prefix+")",
			)
			return
		}
		item, err := prefixModelFromAPI(prefix, statuses.Name)
		if err != nil {
			resp.Diagnostics.AddError("Failed to resolve prefix status", err.Error())
			return
		}

		state.Prefixes = append(state.Prefixes, item)
	}

	tflog.Debug(ctx, "read prefixes", map[string]any{"count": len(state.Prefixes)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
