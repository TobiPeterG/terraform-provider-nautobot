package prefix

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
	_ datasource.DataSource                     = &PrefixDataSource{}
	_ datasource.DataSourceWithConfigure        = &PrefixDataSource{}
	_ datasource.DataSourceWithConfigValidators = &PrefixDataSource{}
)

var prefixSelector = shared.SelectorSpec{NaturalKey: []string{"prefix", "namespace_id"}}

type PrefixDataSource struct {
	client *shared.APIClient
}

type prefixDataSourceModel = prefixItemModel

func NewPrefixDataSource() datasource.DataSource {
	return &PrefixDataSource{}
}

func (d *PrefixDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return prefixSelector.ConfigValidators(ctx)
}

func (d *PrefixDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefix"
}

func (d *PrefixDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a prefix in Nautobot by either its ID or the combination of an exact prefix and namespace UUID.",
		Attributes:  prefixDataAttributes(true),
	}
}

func (d *PrefixDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *PrefixDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data prefixDataSourceModel

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

	idStr := data.ID.ValueString()
	prefixStr := data.Prefix.ValueString()
	namespaceIDStr := data.NamespaceID.ValueString()

	if err := prefixSelector.Validate(idStr, map[string]string{
		"prefix":       prefixStr,
		"namespace_id": namespaceIDStr,
	}); err != nil {
		resp.Diagnostics.AddError("Invalid prefix selector", err.Error())
		return
	}

	var prefix *nb.Prefix

	if idStr != "" {
		// Fetch prefix by ID
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesRetrieve(ctx, idStr).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get prefix by ID",
				shared.HTTPError(err, httpResp),
			)
			return
		}
		prefix = rsp
	} else {
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesList(ctx).
			Prefix([]string{prefixStr}).
			Namespace([]string{namespaceIDStr}).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get prefix by prefix",
				shared.HTTPError(err, httpResp),
			)
			return
		}

		selectorDescription := fmt.Sprintf("prefix %q in namespace %q", prefixStr, namespaceIDStr)
		if err := shared.ExactMatchError("prefix", selectorDescription, len(rsp.Results)); err != nil {
			resp.Diagnostics.AddError("Prefix lookup failed", err.Error())
			return
		}
		prefix = &rsp.Results[0]
	}

	var err error
	data, err = prefixModelFromAPI(prefix, func(statusID string) (string, error) {
		return shared.GetStatusName(ctx, c, statusID)
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve prefix status", err.Error())
		return
	}
	if err := shared.ValidateReturnedObjectID("prefix", idStr, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid prefix data", err.Error())
		return
	}
	resID := data.ID.ValueString()

	tflog.Debug(ctx, "read Prefix", map[string]any{"id": resID, "prefix": data.Prefix.ValueString(), "namespace_id": data.NamespaceID.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
