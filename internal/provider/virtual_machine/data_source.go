package virtual_machine

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
	_ datasource.DataSource                     = &VirtualMachineDataSource{}
	_ datasource.DataSourceWithConfigure        = &VirtualMachineDataSource{}
	_ datasource.DataSourceWithConfigValidators = &VirtualMachineDataSource{}
)

var virtualMachineSelector = shared.SelectorSpec{
	NaturalKey: []string{"name", "cluster_id"},
	Qualifiers: []string{"tenant_id"},
}

type VirtualMachineDataSource struct {
	client *shared.APIClient
}

type virtualMachineDataSourceModel = virtualMachineItemModel

func NewVirtualMachineDataSource() datasource.DataSource {
	return &VirtualMachineDataSource{}
}

func (d *VirtualMachineDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return virtualMachineSelector.ConfigValidators(ctx)
}

func (d *VirtualMachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (d *VirtualMachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific virtual machine in Nautobot by ID or by its exact cluster, tenant, and name natural key.",
		Attributes:  virtualMachineDataAttributes(true),
	}
}

func (d *VirtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *VirtualMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data virtualMachineDataSourceModel

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

	id := data.ID.ValueString()
	vmName := data.Name.ValueString()
	clusterID := data.ClusterID.ValueString()
	tenantID := data.TenantID.ValueString()
	if err := virtualMachineSelector.Validate(id, map[string]string{
		"name":       vmName,
		"cluster_id": clusterID,
		"tenant_id":  tenantID,
	}); err != nil {
		resp.Diagnostics.AddError("Invalid virtual machine selector", err.Error())
		return
	}

	var vm *nb.VirtualMachine
	if id != "" {
		out, httpResp, err := c.VirtualizationAPI.VirtualizationVirtualMachinesRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get virtual machine by ID", shared.HTTPError(err, httpResp))
			return
		}
		vm = out
	} else {
		query := c.VirtualizationAPI.VirtualizationVirtualMachinesList(ctx).
			Name([]string{vmName}).
			Cluster([]string{clusterID})
		selectorDescription := fmt.Sprintf("name %q in cluster %q with no tenant", vmName, clusterID)
		if tenantID != "" {
			query = query.Tenant([]string{tenantID})
			selectorDescription = fmt.Sprintf("name %q in cluster %q and tenant %q", vmName, clusterID, tenantID)
		} else {
			query = query.TenantIsnull(true)
		}
		rsp, httpResp, err := query.Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get virtual machine by natural key", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("virtual machine", selectorDescription, len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Virtual machine lookup failed", err.Error())
			return
		}
		vm = &rsp.Results[0]
	}

	data, err := virtualMachineModelFromAPI(vm, func(id string) (string, error) {
		return shared.GetStatusName(ctx, c, id)
	})
	if err != nil {
		resp.Diagnostics.AddError("Invalid virtual machine data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("virtual machine", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid virtual machine data", err.Error())
		return
	}

	tflog.Debug(ctx, "read virtual machine", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
