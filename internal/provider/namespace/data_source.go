package namespace

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
	_ datasource.DataSource                     = &NamespaceDataSource{}
	_ datasource.DataSourceWithConfigure        = &NamespaceDataSource{}
	_ datasource.DataSourceWithConfigValidators = &NamespaceDataSource{}
)

var namespaceSelector = shared.SelectorSpec{NaturalKey: []string{"name"}}

type NamespaceDataSource struct {
	client *shared.APIClient
}

type namespaceDataSourceModel = namespaceItemModel

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

func (d *NamespaceDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return namespaceSelector.ConfigValidators(ctx)
}

func (d *NamespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific IPAM namespace in Nautobot by ID or exact name.",
		Attributes:  namespaceDataAttributes(true),
	}
}

func (d *NamespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data namespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	id, name := data.ID.ValueString(), data.Name.ValueString()
	if err := namespaceSelector.Validate(id, map[string]string{"name": name}); err != nil {
		resp.Diagnostics.AddError("Invalid namespace selector", err.Error())
		return
	}

	var n *nb.Namespace
	if id != "" {
		out, httpResp, err := d.client.Client.IpamAPI.IpamNamespacesRetrieve(ctx, id).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get namespace by ID", shared.HTTPError(err, httpResp))
			return
		}
		n = out
	} else {
		rsp, httpResp, err := d.client.Client.IpamAPI.IpamNamespacesList(ctx).Name([]string{name}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get namespace by name", shared.HTTPError(err, httpResp))
			return
		}
		if err := shared.ExactMatchError("namespace", fmt.Sprintf("name %q", name), len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Namespace lookup failed", err.Error())
			return
		}
		n = &rsp.Results[0]
	}

	data, err := namespaceModelFromAPI(n)
	if err != nil {
		resp.Diagnostics.AddError("Invalid namespace data", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("namespace", id, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid namespace data", err.Error())
		return
	}

	tflog.Debug(ctx, "read namespace", map[string]any{"id": data.ID.ValueString(), "name": data.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
