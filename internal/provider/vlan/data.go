package vlan

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type vlanItemModel struct {
	ID          types.String `tfsdk:"id"`
	VID         types.Int64  `tfsdk:"vid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	VLANGroupID types.String `tfsdk:"vlan_group_id"`
	Status      types.String `tfsdk:"status"`
	TenantID    types.String `tfsdk:"tenant_id"`
	RoleID      types.String `tfsdk:"role_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	PrefixCount types.Int64  `tfsdk:"prefix_count"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func vlanDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":            dsschema.StringAttribute{Description: "The UUID of the VLAN.", Computed: true},
		"vid":           dsschema.Int64Attribute{Description: "The ID (VID) of the VLAN.", Computed: true},
		"name":          dsschema.StringAttribute{Description: "The exact name of the VLAN.", Computed: true},
		"description":   dsschema.StringAttribute{Description: "Description of the VLAN.", Computed: true},
		"vlan_group_id": dsschema.StringAttribute{Description: "The ID of the VLAN group. For name lookup, omit it to match only ungrouped VLANs.", Computed: true},
		"status":        dsschema.StringAttribute{Description: "The status of the VLAN (name).", Computed: true},
		"tenant_id":     dsschema.StringAttribute{Description: "The ID of the tenant associated with the VLAN.", Computed: true},
		"role_id":       dsschema.StringAttribute{Description: "The ID of the role associated with the VLAN.", Computed: true},
		"tags_ids":      dsschema.ListAttribute{Description: "The IDs of the tags associated with the VLAN.", Computed: true, ElementType: types.StringType},
		"created":       dsschema.StringAttribute{Description: "The creation date of the VLAN (RFC3339).", Computed: true},
		"last_updated":  dsschema.StringAttribute{Description: "The last update date of the VLAN (RFC3339).", Computed: true},
		"prefix_count":  dsschema.Int64Attribute{Description: "Number of prefixes associated with this VLAN.", Computed: true},
		"display":       dsschema.StringAttribute{Description: "Human-friendly display value.", Computed: true},
		"url":           dsschema.StringAttribute{Description: "API URL of the VLAN.", Computed: true},
		"natural_slug":  dsschema.StringAttribute{Description: "Natural slug for the VLAN.", Computed: true},
		"notes_url":     dsschema.StringAttribute{Description: "Notes URL for the VLAN.", Computed: true},
	}
	if selectable {
		for _, name := range []string{"id", "name", "vlan_group_id"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = true
			attributes[name] = attribute
		}
	}
	return attributes
}

func vlanModelFromAPI(vlan *nb.VLAN, statusName func(string) (string, error)) (vlanItemModel, error) {
	if vlan == nil || vlan.Id == nil || *vlan.Id == "" {
		return vlanItemModel{}, fmt.Errorf("VLAN returned no id")
	}
	status := ""
	if vlan.Status.Id != nil && vlan.Status.Id.String != nil && *vlan.Status.Id.String != "" {
		var err error
		status, err = statusName(*vlan.Status.Id.String)
		if err != nil {
			return vlanItemModel{}, fmt.Errorf("resolve VLAN status: %w", err)
		}
	}
	prefixCount := int64(0)
	if vlan.PrefixCount != nil {
		prefixCount = int64(*vlan.PrefixCount)
	}
	return vlanItemModel{
		ID: types.StringValue(*vlan.Id), VID: types.Int64Value(int64(vlan.Vid)), Name: types.StringValue(vlan.Name),
		Description: types.StringValue(shared.DerefString(vlan.Description)), VLANGroupID: shared.NullableReferenceID(vlan.VlanGroup),
		Status: types.StringValue(status), TenantID: shared.NullableReferenceID(vlan.Tenant), RoleID: shared.NullableReferenceID(vlan.Role),
		TagsIDs: shared.ReferenceIDs(vlan.Tags), Created: shared.NullableTimeValue(vlan.Created), LastUpdated: shared.NullableTimeValue(vlan.LastUpdated),
		PrefixCount: types.Int64Value(prefixCount), Display: types.StringValue(vlan.Display), URL: types.StringValue(vlan.Url),
		NaturalSlug: types.StringValue(vlan.NaturalSlug), NotesURL: types.StringValue(vlan.NotesUrl),
	}, nil
}
