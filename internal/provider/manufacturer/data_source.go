package manufacturer

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
	_ datasource.DataSource                     = &ManufacturerDataSource{}
	_ datasource.DataSourceWithConfigure        = &ManufacturerDataSource{}
	_ datasource.DataSourceWithConfigValidators = &ManufacturerDataSource{}
)

var manufacturerSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type ManufacturerDataSource struct {
	client *shared.APIClient
}

type manufacturerDataSourceModel = manufacturerItemModel

func NewManufacturerDataSource() datasource.DataSource {
	return &ManufacturerDataSource{}
}

func (d *ManufacturerDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return manufacturerSelector.ConfigValidators(ctx)
}

func (d *ManufacturerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturer"
}

func (d *ManufacturerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific manufacturer in Nautobot by ID or exact name.",
		Attributes:  manufacturerDataAttributes(true),
	}
}

func (d *ManufacturerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ManufacturerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data manufacturerDataSourceModel

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
	if err := manufacturerSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid manufacturer selector", err.Error())
		return
	}

	var m *nb.Manufacturer
	if id != "" {
		out, httpResp, err := c.DcimAPI.DcimManufacturersRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get manufacturer by ID", shared.HTTPError(err, httpResp))
			return
		}
		m = out
	} else {
		rsp, httpResp, err := c.DcimAPI.DcimManufacturersList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get manufacturer by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("manufacturer", fmt.Sprintf("name %q", name), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Manufacturer lookup failed", err.Error())
			return
		}
		m = &rsp.Results[0]
	}

	model, err := manufacturerModelFromAPI(m)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid manufacturer data",
			err.Error(),
		)
		return
	}
	if err := shared.ValidateReturnedObjectID("manufacturer", id, model.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid manufacturer data", err.Error())
		return
	}

	tflog.Debug(ctx, "read manufacturer", map[string]any{"id": model.ID.ValueString(), "name": model.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
