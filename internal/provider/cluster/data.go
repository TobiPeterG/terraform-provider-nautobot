package cluster

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type clusterItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ClusterTypeID  types.String `tfsdk:"cluster_type_id"`
	ClusterGroupID types.String `tfsdk:"cluster_group_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	LocationID     types.String `tfsdk:"location_id"`
	TagsIDs        types.List   `tfsdk:"tags_ids"`
	Comments       types.String `tfsdk:"comments"`
	Created        types.String `tfsdk:"created"`
	LastUpdated    types.String `tfsdk:"last_updated"`
	Display        types.String `tfsdk:"display"`
	URL            types.String `tfsdk:"url"`
	NaturalSlug    types.String `tfsdk:"natural_slug"`
	NotesURL       types.String `tfsdk:"notes_url"`
	DeviceCount    types.Int64  `tfsdk:"device_count"`
	VMCount        types.Int64  `tfsdk:"virtual_machine_count"`
}

func clusterDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":                    dsschema.StringAttribute{Description: "The UUID of the cluster.", Computed: true},
		"name":                  dsschema.StringAttribute{Description: "The name of the cluster.", Computed: true},
		"cluster_type_id":       dsschema.StringAttribute{Description: "The ID of the cluster type.", Computed: true},
		"cluster_group_id":      dsschema.StringAttribute{Description: "The ID of the cluster group.", Computed: true},
		"tenant_id":             dsschema.StringAttribute{Description: "The ID of the tenant associated with the cluster.", Computed: true},
		"location_id":           dsschema.StringAttribute{Description: "The ID of the location associated with the cluster.", Computed: true},
		"tags_ids":              dsschema.ListAttribute{Description: "The IDs of the tags associated with the cluster.", Computed: true, ElementType: types.StringType},
		"comments":              dsschema.StringAttribute{Description: "Comments or notes about the cluster.", Computed: true},
		"created":               dsschema.StringAttribute{Description: "The creation date of the cluster (RFC3339).", Computed: true},
		"last_updated":          dsschema.StringAttribute{Description: "The last update date of the cluster (RFC3339).", Computed: true},
		"display":               dsschema.StringAttribute{Description: "Human-friendly display value for the cluster.", Computed: true},
		"url":                   dsschema.StringAttribute{Description: "API URL of the cluster.", Computed: true},
		"natural_slug":          dsschema.StringAttribute{Description: "Natural slug for the cluster.", Computed: true},
		"notes_url":             dsschema.StringAttribute{Description: "Notes URL for the cluster.", Computed: true},
		"device_count":          dsschema.Int64Attribute{Description: "Number of devices associated with the cluster.", Computed: true},
		"virtual_machine_count": dsschema.Int64Attribute{Description: "Number of virtual machines in the cluster.", Computed: true},
	}
	if selectable {
		for _, name := range []string{"id", "name"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = true
			attributes[name] = attribute
		}
	}
	return attributes
}

func clusterModelFromAPI(cluster *nb.Cluster) (clusterItemModel, error) {
	if cluster == nil || cluster.Id == nil || *cluster.Id == "" {
		return clusterItemModel{}, fmt.Errorf("cluster returned no id")
	}
	clusterTypeID := types.StringValue("")
	if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
		clusterTypeID = types.StringValue(*cluster.ClusterType.Id.String)
	}
	deviceCount, virtualMachineCount := int64(0), int64(0)
	if cluster.DeviceCount != nil {
		deviceCount = int64(*cluster.DeviceCount)
	}
	if cluster.VirtualmachineCount != nil {
		virtualMachineCount = int64(*cluster.VirtualmachineCount)
	}
	return clusterItemModel{
		ID:             types.StringValue(*cluster.Id),
		Name:           types.StringValue(cluster.Name),
		ClusterTypeID:  clusterTypeID,
		ClusterGroupID: shared.NullableReferenceID(cluster.ClusterGroup),
		TenantID:       shared.NullableReferenceID(cluster.Tenant),
		LocationID:     shared.NullableReferenceID(cluster.Location),
		TagsIDs:        shared.ReferenceIDs(cluster.Tags),
		Comments:       types.StringValue(shared.DerefString(cluster.Comments)),
		Created:        shared.NullableTimeValue(cluster.Created),
		LastUpdated:    shared.NullableTimeValue(cluster.LastUpdated),
		Display:        types.StringValue(cluster.Display),
		URL:            types.StringValue(cluster.Url),
		NaturalSlug:    types.StringValue(cluster.NaturalSlug),
		NotesURL:       types.StringValue(cluster.NotesUrl),
		DeviceCount:    types.Int64Value(deviceCount),
		VMCount:        types.Int64Value(virtualMachineCount),
	}, nil
}
