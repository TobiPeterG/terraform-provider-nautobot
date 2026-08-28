package vlan

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

// vlanGroupDataModel and vlanGroupDataAttributes are shared by the singular and
// plural data sources. Keeping the API-to-state mapping here prevents the two
// data sources from drifting as Nautobot adds fields.
type vlanGroupDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Range       types.String `tfsdk:"range"`
	LocationID  types.String `tfsdk:"location_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`
	VLANCount   types.Int64  `tfsdk:"vlan_count"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func vlanGroupDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "VLAN group UUID.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "VLAN group name.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "VLAN group description.", Computed: true},
		"range":        dsschema.StringAttribute{Description: "Permitted VLAN IDs as a comma-separated list of ranges.", Computed: true},
		"location_id":  dsschema.StringAttribute{Description: "UUID of the location associated with the VLAN group.", Computed: true},
		"tags_ids":     dsschema.ListAttribute{Description: "UUIDs of tags associated with the VLAN group.", Computed: true, ElementType: types.StringType},
		"vlan_count":   dsschema.Int64Attribute{Description: "Number of VLANs in the VLAN group.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "VLAN group creation date (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "VLAN group last update date (RFC3339).", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human-friendly display value for the VLAN group.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "API URL of the VLAN group.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the VLAN group.", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the VLAN group.", Computed: true},
	}
	if selectable {
		attributes["id"] = dsschema.StringAttribute{
			Description: "VLAN group UUID. Provide either `id` or `name`.",
			Optional:    true,
			Computed:    true,
		}
		attributes["name"] = dsschema.StringAttribute{
			Description: "Exact VLAN group name. Provide either `id` or `name`.",
			Optional:    true,
			Computed:    true,
		}
	}
	return attributes
}

func vlanGroupDataFromAPI(group *nb.VLANGroup) (vlanGroupDataModel, error) {
	if group == nil || group.Id == nil || *group.Id == "" {
		return vlanGroupDataModel{}, fmt.Errorf("VLAN group response has no id")
	}

	return vlanGroupDataModel{
		ID:          types.StringValue(*group.Id),
		Name:        types.StringValue(group.Name),
		Description: types.StringValue(shared.DerefString(group.Description)),
		Range:       types.StringValue(shared.DerefString(group.Range)),
		LocationID:  shared.NullableReferenceID(group.Location),
		TagsIDs:     shared.ReferenceIDs(group.Tags),
		VLANCount:   types.Int64Value(int64(group.GetVlanCount())),
		Created:     shared.NullableTimeValue(group.Created),
		LastUpdated: shared.NullableTimeValue(group.LastUpdated),
		Display:     types.StringValue(group.Display),
		URL:         types.StringValue(group.Url),
		NaturalSlug: types.StringValue(group.NaturalSlug),
		NotesURL:    types.StringValue(group.NotesUrl),
	}, nil
}
