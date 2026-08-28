package cluster_type

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

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

func clusterTypeModelFromAPI(clusterType *nb.ClusterType) (clusterTypeItemModel, error) {
	if clusterType == nil || clusterType.Id == nil || *clusterType.Id == "" {
		return clusterTypeItemModel{}, fmt.Errorf("cluster type returned no id")
	}
	return clusterTypeItemModel{
		ID:          types.StringValue(*clusterType.Id),
		Display:     types.StringValue(clusterType.Display),
		URL:         types.StringValue(clusterType.Url),
		NaturalSlug: types.StringValue(clusterType.NaturalSlug),
		Name:        types.StringValue(clusterType.Name),
		Description: types.StringValue(shared.DerefString(clusterType.Description)),
		Created:     shared.NullableTimeValue(clusterType.Created),
		LastUpdated: shared.NullableTimeValue(clusterType.LastUpdated),
		NotesURL:    types.StringValue(clusterType.NotesUrl),
	}, nil
}

func clusterTypeDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "The UUID of the cluster type.", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human-friendly display value for the cluster type.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "URL of the cluster type.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the cluster type.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "The name of the cluster type.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "The description of the cluster type.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "The date the cluster type was created (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "The date the cluster type was last updated (RFC3339).", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the cluster type.", Computed: true},
	}
	if selectable {
		for _, name := range []string{"id", "name"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = true
			attribute.Computed = true
			attributes[name] = attribute
		}
	}
	return attributes
}
