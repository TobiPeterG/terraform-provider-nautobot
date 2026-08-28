package manufacturer

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &ManufacturersDataSource{}
	_ datasource.DataSourceWithConfigure = &ManufacturersDataSource{}
)

type ManufacturersDataSource struct {
	client *shared.APIClient
}

type manufacturersDataSourceModel struct {
	Manufacturers []manufacturerItemModel `tfsdk:"manufacturers"`
}

func NewManufacturersDataSource() datasource.DataSource {
	return &ManufacturersDataSource{}
}

func (d *ManufacturersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturers"
}

func (d *ManufacturersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all manufacturers in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"manufacturers": dsschema.ListNestedAttribute{
				Description: "List of manufacturers.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: manufacturerDataAttributes(false),
				},
			},
		},
	}
}

func (d *ManufacturersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ManufacturersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state manufacturersDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	state.Manufacturers = make([]manufacturerItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.Manufacturer], error) {
		rsp, httpResp, err := c.DcimAPI.
			DcimManufacturersList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.Manufacturer]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.Manufacturer]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list manufacturers", err.Error())
		return
	}

	for _, m := range results {
		item, err := manufacturerModelFromAPI(&m)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid manufacturer data",
				err.Error(),
			)
			return
		}
		state.Manufacturers = append(state.Manufacturers, item)
	}

	tflog.Debug(ctx, "read manufacturers", map[string]any{"count": len(state.Manufacturers)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
