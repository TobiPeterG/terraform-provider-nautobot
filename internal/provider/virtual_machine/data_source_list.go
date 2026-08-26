package virtual_machine

import (
	"context"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &VirtualMachinesDataSource{}
	_ datasource.DataSourceWithConfigure = &VirtualMachinesDataSource{}
)

type VirtualMachinesDataSource struct {
	client *shared.APIClient
}

type virtualMachinesDataSourceModel struct {
	VirtualMachines []virtualMachineItemModel `tfsdk:"virtual_machines"`
}

func NewVirtualMachinesDataSource() datasource.DataSource {
	return &VirtualMachinesDataSource{}
}

func (d *VirtualMachinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

func (d *VirtualMachinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all virtual machines in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"virtual_machines": dsschema.ListNestedAttribute{
				Description: "List of virtual machines.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: virtualMachineDataAttributes(false),
				},
			},
		},
	}
}

func (d *VirtualMachinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VirtualMachinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachinesDataSourceModel

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

	state.VirtualMachines = make([]virtualMachineItemModel, 0)
	results, err := shared.CollectPages(func(limit, offset int32) (shared.Page[nb.VirtualMachine], error) {
		rsp, httpResp, err := c.VirtualizationAPI.
			VirtualizationVirtualMachinesList(ctx).
			Limit(limit).
			Offset(offset).
			Sort("id").
			Execute()
		if err != nil {
			return shared.Page[nb.VirtualMachine]{}, shared.HTTPErrorAsError(err, httpResp)
		}
		return shared.Page[nb.VirtualMachine]{Items: rsp.Results, HasNext: rsp.Next.IsSet() && rsp.Next.Get() != nil && *rsp.Next.Get() != ""}, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list virtual machines", err.Error())
		return
	}

	for _, vm := range results {
		item, err := virtualMachineModelFromAPI(&vm, statuses.Name)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid virtual machine data",
				err.Error()+" (name: "+vm.Name+")",
			)
			return
		}
		state.VirtualMachines = append(state.VirtualMachines, item)
	}

	tflog.Debug(ctx, "read virtual machines", map[string]any{"count": len(state.VirtualMachines)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
