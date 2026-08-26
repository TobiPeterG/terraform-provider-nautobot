package tenant_group

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
	_ datasource.DataSource                     = &TenantGroupDataSource{}
	_ datasource.DataSourceWithConfigure        = &TenantGroupDataSource{}
	_ datasource.DataSourceWithConfigValidators = &TenantGroupDataSource{}
)

var tenantGroupSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type TenantGroupDataSource struct {
	client *shared.APIClient
}

type tenantGroupDataSourceModel = tenantGroupItemModel

func NewTenantGroupDataSource() datasource.DataSource {
	return &TenantGroupDataSource{}
}

func (d *TenantGroupDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return tenantGroupSelector.ConfigValidators(ctx)
}

func (d *TenantGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_group"
}

func (d *TenantGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific tenant group in Nautobot by ID or exact name.",
		Attributes:  tenantGroupDataAttributes(true),
	}
}

func (d *TenantGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *TenantGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tenantGroupDataSourceModel

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
	if err := tenantGroupSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid tenant group selector", err.Error())
		return
	}

	var m *nb.TenantGroup
	if id != "" {
		out, httpResp, err := c.TenancyAPI.TenancyTenantGroupsRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get tenant group by ID", shared.HTTPError(err, httpResp))
			return
		}
		m = out
	} else {
		rsp, httpResp, err := c.TenancyAPI.TenancyTenantGroupsList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get tenant group by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("tenant group", fmt.Sprintf("name %q", name), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Tenant group lookup failed", err.Error())
			return
		}
		m = &rsp.Results[0]
	}

	data, err := tenantGroupModelFromAPI(m)
	if err != nil {
		resp.Diagnostics.AddError("Invalid tenant group data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("tenant group", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid tenant group data", err.Error())
		return
	}

	tflog.Debug(ctx, "read tenant group", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
