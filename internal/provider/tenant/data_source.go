package tenant

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
	_ datasource.DataSource                     = &TenantDataSource{}
	_ datasource.DataSourceWithConfigure        = &TenantDataSource{}
	_ datasource.DataSourceWithConfigValidators = &TenantDataSource{}
)

var tenantSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type TenantDataSource struct {
	client *shared.APIClient
}

type tenantDataSourceModel = tenantItemModel

func NewTenantDataSource() datasource.DataSource {
	return &TenantDataSource{}
}

func (d *TenantDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return tenantSelector.ConfigValidators(ctx)
}

func (d *TenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *TenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific tenant in Nautobot by ID or exact name.",
		Attributes:  tenantDataAttributes(true),
	}
}

func (d *TenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *TenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tenantDataSourceModel

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
	if err := tenantSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid tenant selector", err.Error())
		return
	}

	var m *nb.Tenant
	if id != "" {
		out, httpResp, err := c.TenancyAPI.TenancyTenantsRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get tenant by ID", shared.HTTPError(err, httpResp))
			return
		}
		m = out
	} else {
		rsp, httpResp, err := c.TenancyAPI.TenancyTenantsList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get tenant by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("tenant", fmt.Sprintf("name %q", name), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Tenant lookup failed", err.Error())
			return
		}
		m = &rsp.Results[0]
	}

	data, err := tenantModelFromAPI(m)
	if err != nil {
		resp.Diagnostics.AddError("Invalid tenant data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("tenant", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid tenant data", err.Error())
		return
	}

	tflog.Debug(ctx, "read tenant", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
