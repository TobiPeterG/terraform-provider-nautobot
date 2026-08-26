package graphql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &GraphQLDataSource{}
	_ datasource.DataSourceWithConfigure = &GraphQLDataSource{}
)

type GraphQLDataSource struct {
	client *shared.APIClient
}

type graphQLDataSourceModel struct {
	Query types.String `tfsdk:"query"`
	Data  types.String `tfsdk:"data"`
}

func NewGraphQLDataSource() datasource.DataSource {
	return &GraphQLDataSource{}
}

func (d *GraphQLDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graphql"
}

func (d *GraphQLDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Provide an interface to make GraphQL calls to Nautobot as a flexible data source.",
		Attributes: map[string]dsschema.Attribute{
			"query": dsschema.StringAttribute{
				Description: "The GraphQL query that will be sent to Nautobot.",
				Required:    true,
			},
			"data": dsschema.StringAttribute{
				Description: "The data returned by the GraphQL query (JSON-encoded GraphQL `data` field).",
				Computed:    true,
			},
		},
	}
}

func (d *GraphQLDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = shared.ConfigureAPIClient(req.ProviderData, &resp.Diagnostics)
}

func (d *GraphQLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data graphQLDataSourceModel

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

	queryStr := data.Query.ValueString()
	tflog.Debug(ctx, "executing GraphQL query", map[string]any{"query_length": len(queryStr)})

	body := nb.GraphQLAPIRequest{
		Query: queryStr,
	}

	gqlResp, httpResp, err := c.GraphqlAPI.
		GraphqlCreate(ctx).
		GraphQLAPIRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute GraphQL query",
			shared.HTTPError(err, httpResp),
		)
		return
	}
	if gqlResp != nil {
		if errors, ok := gqlResp.AdditionalProperties["errors"]; ok && graphQLResponseHasErrors(errors) {
			details, marshalErr := json.Marshal(errors)
			if marshalErr != nil {
				details = []byte(fmt.Sprintf("%v", errors))
			}
			resp.Diagnostics.AddError(
				"GraphQL query returned errors",
				string(details),
			)
			return
		}
	}

	var raw string
	if gqlResp != nil {
		if gqlResp.Data != nil {
			if b, err := json.Marshal(gqlResp.Data); err == nil {
				raw = string(b)
			} else {
				resp.Diagnostics.AddError(
					"Failed to marshal GraphQL data",
					err.Error(),
				)
				return
			}
		} else {
			raw = "null"
		}
	} else {
		raw = "null"
	}

	data.Data = types.StringValue(raw)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func graphQLResponseHasErrors(value any) bool {
	switch errors := value.(type) {
	case nil:
		return false
	case []any:
		return len(errors) > 0
	default:
		return true
	}
}
