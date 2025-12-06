package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ClusterTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &ClusterTypesDataSource{}
)

type ClusterTypesDataSource struct {
	client *APIClient
}

type clusterTypeItemModel struct {
	ID          types.String `tfsdk:"id"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

type clusterTypesDataSourceModel struct {
	ClusterTypes []clusterTypeItemModel `tfsdk:"cluster_types"`
}

func NewClusterTypesDataSource() datasource.DataSource {
	return &ClusterTypesDataSource{}
}

func (d *ClusterTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_types"
}

func (d *ClusterTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about cluster types in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"cluster_types": dsschema.ListNestedAttribute{
				Description: "List of cluster types.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The UUID of the cluster type.",
							Computed:    true,
						},
						"display": dsschema.StringAttribute{
							Description: "Human-friendly display value for the cluster type.",
							Computed:    true,
						},
						"url": dsschema.StringAttribute{
							Description: "URL of the cluster type.",
							Computed:    true,
						},
						"natural_slug": dsschema.StringAttribute{
							Description: "Natural slug for the cluster type.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the cluster type.",
							Computed:    true,
						},
						"description": dsschema.StringAttribute{
							Description: "The description of the cluster type.",
							Computed:    true,
						},
						"created": dsschema.StringAttribute{
							Description: "The date the cluster type was created (RFC3339).",
							Computed:    true,
						},
						"last_updated": dsschema.StringAttribute{
							Description: "The date the cluster type was last updated (RFC3339).",
							Computed:    true,
						},
						"notes_url": dsschema.StringAttribute{
							Description: "Notes URL for the cluster type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ClusterTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ClusterTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterTypesDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	rsp, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesList(ctx).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster types list",
			httpErr(err, httpResp),
		)
		return
	}

	results := rsp.Results
	state.ClusterTypes = make([]clusterTypeItemModel, 0, len(results))

	for _, ct := range results {
		var item clusterTypeItemModel

		idStr := ""
		if ct.Id != nil {
			idStr = *ct.Id
		}
		item.ID = types.StringValue(idStr)

		createdStr := ""
		if ct.Created.IsSet() && ct.Created.Get() != nil {
			createdStr = ct.Created.Get().Format(time.RFC3339)
		}
		lastUpdatedStr := ""
		if ct.LastUpdated.IsSet() && ct.LastUpdated.Get() != nil {
			lastUpdatedStr = ct.LastUpdated.Get().Format(time.RFC3339)
		}

		descStr := ""
		if ct.Description != nil {
			descStr = *ct.Description
		}

		item.Display = types.StringValue(ct.Display)
		item.URL = types.StringValue(ct.Url)
		item.NaturalSlug = types.StringValue(ct.NaturalSlug)
		item.Name = types.StringValue(ct.Name)
		item.Description = types.StringValue(descStr)
		item.Created = types.StringValue(createdStr)
		item.LastUpdated = types.StringValue(lastUpdatedStr)
		item.NotesURL = types.StringValue(ct.NotesUrl)

		state.ClusterTypes = append(state.ClusterTypes, item)
	}

	tflog.Debug(ctx, "read cluster types", map[string]any{"count": len(state.ClusterTypes)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
