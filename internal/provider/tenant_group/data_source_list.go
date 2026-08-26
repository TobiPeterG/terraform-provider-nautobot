package tenant_group

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &TenantGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantGroupsDataSource{}
)

type TenantGroupsDataSource struct {
	client *shared.APIClient
}

type tenantGroupsDataSourceModel struct {
	TenantGroups []tenantGroupItemModel `tfsdk:"tenant_groups"`
}

func NewTenantGroupsDataSource() datasource.DataSource {
	return &TenantGroupsDataSource{}
}

func (d *TenantGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_groups"
}

func (d *TenantGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all tenant groups in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"tenant_groups": dsschema.ListNestedAttribute{
				Description:  "List of tenant groups.",
				Computed:     true,
				NestedObject: dsschema.NestedAttributeObject{Attributes: tenantGroupDataAttributes(false)},
			},
		},
	}
}

func (d *TenantGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *TenantGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantGroupsDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	state.TenantGroups = make([]tenantGroupItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.TenantGroup], error) {
		rsp, httpResp, err := c.TenancyAPI.
			TenancyTenantGroupsList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.TenantGroup]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.TenantGroup]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list tenant groups", err.Error())
		return
	}

	for _, m := range results {
		item, err := tenantGroupModelFromAPI(&m)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid tenant group data",
				err.Error()+" (name: "+m.Name+")",
			)
			return
		}
		state.TenantGroups = append(state.TenantGroups, item)
	}

	tflog.Debug(ctx, "read tenant groups", map[string]any{"count": len(state.TenantGroups)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
