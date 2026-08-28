package tenant

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &TenantsDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantsDataSource{}
)

type TenantsDataSource struct {
	client *shared.APIClient
}

type tenantsDataSourceModel struct {
	Tenants []tenantItemModel `tfsdk:"tenants"`
}

func NewTenantsDataSource() datasource.DataSource {
	return &TenantsDataSource{}
}

func (d *TenantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenants"
}

func (d *TenantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all tenants in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"tenants": dsschema.ListNestedAttribute{
				Description: "List of tenants.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: tenantDataAttributes(false),
				},
			},
		},
	}
}

func (d *TenantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *TenantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantsDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	state.Tenants = make([]tenantItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.Tenant], error) {
		rsp, httpResp, err := c.TenancyAPI.
			TenancyTenantsList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.Tenant]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.Tenant]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list tenants", err.Error())
		return
	}

	for _, m := range results {
		item, err := tenantModelFromAPI(&m)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid tenant data",
				err.Error()+" (name: "+m.Name+")",
			)
			return
		}
		state.Tenants = append(state.Tenants, item)
	}

	tflog.Debug(ctx, "read tenants", map[string]any{"count": len(state.Tenants)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
