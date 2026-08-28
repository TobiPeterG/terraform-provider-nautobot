package ip_address_range

import (
	"context"
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource                     = &IPAddressRangeDataSource{}
	_ datasource.DataSourceWithConfigure        = &IPAddressRangeDataSource{}
	_ datasource.DataSourceWithConfigValidators = &IPAddressRangeDataSource{}
)

var ipAddressRangeSelector = shared.SelectorSpec{
	NaturalKey: []string{"start_address", "end_address", "namespace_id"},
}

type IPAddressRangeDataSource struct {
	client *shared.APIClient
}

type ipAddressRangeDataSourceModel = ipAddressRangeItemModel

func NewIPAddressRangeDataSource() datasource.DataSource {
	return &IPAddressRangeDataSource{}
}
func (d *IPAddressRangeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}

func (d *IPAddressRangeDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return ipAddressRangeSelector.ConfigValidators(ctx)
}
func (d *IPAddressRangeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func ipAddressRangeDataAttributes(selectable bool) map[string]dsschema.Attribute {
	computedString := func(description string) dsschema.StringAttribute {
		return dsschema.StringAttribute{Computed: true, Description: description}
	}
	attributes := map[string]dsschema.Attribute{
		"id": dsschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Range UUID. Provide either `id`, or all of `start_address`, `end_address`, and `namespace_id`.",
		},
		"start_address": dsschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "First address in the range (inclusive). Provide together with `end_address` and `namespace_id` when `id` is not used.",
		},
		"end_address": dsschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Last address in the range (inclusive). Provide together with `start_address` and `namespace_id` when `id` is not used.",
		},
		"namespace_id": dsschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Namespace UUID. Provide together with `start_address` and `end_address` when `id` is not used.",
		},
		"parent_id":         computedString("UUID of the containing parent prefix."),
		"name":              computedString("Range name."),
		"description":       computedString("Range description."),
		"count_as_utilized": dsschema.BoolAttribute{Computed: true, Description: "Whether the range counts as fully utilized."},
		"is_exclusive":      dsschema.BoolAttribute{Computed: true, Description: "Whether individual addresses are prohibited."},
		"status":            computedString("Range status name."),
		"role_id":           computedString("Associated role UUID."),
		"tenant_id":         computedString("Associated tenant UUID."),
		"tags_ids":          dsschema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Associated tag UUIDs."},
		"start_host":        computedString("Normalized first host."),
		"end_host":          computedString("Normalized last host."),
		"size":              dsschema.Int64Attribute{Computed: true, Description: "Number of addresses."},
		"ip_version":        dsschema.Int64Attribute{Computed: true, Description: "IP version."},
		"created":           computedString("Creation timestamp."),
		"last_updated":      computedString("Last update timestamp."),
		"display":           computedString("Display value."),
		"url":               computedString("API URL."),
		"natural_slug":      computedString("Natural slug."),
		"notes_url":         computedString("Notes API URL."),
	}
	if !selectable {
		pluralDescriptions := map[string]string{
			"id":            "IP address range UUID.",
			"start_address": "First address in the range (inclusive).",
			"end_address":   "Last address in the range (inclusive).",
			"namespace_id":  "Namespace UUID inherited from the parent prefix.",
		}
		for name, description := range pluralDescriptions {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = false
			attribute.Description = description
			attributes[name] = attribute
		}
	}
	return attributes
}

func (d *IPAddressRangeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves an IP address range in Nautobot by ID or by exact boundaries and namespace.",
		Attributes:  ipAddressRangeDataAttributes(true),
	}
}

func (d *IPAddressRangeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ipAddressRangeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, start, end, namespace := data.ID.ValueString(), data.StartAddress.ValueString(), data.EndAddress.ValueString(), data.NamespaceID.ValueString()
	if err := ipAddressRangeSelector.Validate(id, map[string]string{
		"start_address": start,
		"end_address":   end,
		"namespace_id":  namespace,
	}); err != nil {
		resp.Diagnostics.AddError("Invalid IP address range selector", err.Error())
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}
	var outID string
	if id != "" {
		outID = id
	} else {
		list, httpResp, err := d.client.Client.IpamAPI.IpamIpAddressRangesList(ctx).StartAddress([]string{start}).EndAddress([]string{end}).Namespace([]string{namespace}).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to find IP address range", shared.HTTPError(err, httpResp))
			return
		}
		selectorDescription := fmt.Sprintf("bounds %q-%q in namespace %q", start, end, namespace)
		if err := shared.ExactMatchError("IP address range", selectorDescription, len(list.Results)); err != nil {
			resp.Diagnostics.AddError("IP address range lookup failed", err.Error())
			return
		}
		if list.Results[0].Id == nil {
			resp.Diagnostics.AddError("Invalid API response", "matched range returned no id")
			return
		}
		outID = *list.Results[0].Id
	}
	out, httpResp, err := d.client.Client.IpamAPI.IpamIpAddressRangesRetrieve(ctx, outID).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve IP address range", shared.HTTPError(err, httpResp))
		return
	}
	model, diags := d.buildModel(ctx, out)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := shared.ValidateReturnedObjectID("IP address range", outID, model.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid IP address range data", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (d *IPAddressRangeDataSource) buildModel(ctx context.Context, out *nb.IPAddressRange) (ipAddressRangeDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model, err := ipAddressRangeModelFromAPI(out)
	if err != nil {
		diags.AddError("Invalid IP address range data", err.Error())
		return ipAddressRangeDataSourceModel{}, diags
	}

	statusName := ""
	if out.Status.Id != nil && out.Status.Id.String != nil && *out.Status.Id.String != "" {
		name, err := shared.GetStatusName(ctx, d.client.Client, *out.Status.Id.String)
		if err != nil {
			diags.AddError("Failed to resolve IP address range status", err.Error())
			return ipAddressRangeDataSourceModel{}, diags
		}
		statusName = name
	}
	model.Status = types.StringValue(statusName)

	parentID := model.ParentID.ValueString()
	if parentID != "" {
		parent, httpResp, err := d.client.Client.IpamAPI.IpamPrefixesRetrieve(ctx, parentID).Execute()
		if err != nil {
			diags.AddError("Failed to resolve IP address range namespace", shared.HTTPError(err, httpResp))
			return model, diags
		}
		if parent.Namespace != nil && parent.Namespace.Id != nil && parent.Namespace.Id.String != nil {
			model.NamespaceID = types.StringValue(*parent.Namespace.Id.String)
		}
	}

	return model, diags
}
