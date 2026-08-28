package manufacturer

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type manufacturerItemModel struct {
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

func manufacturerModelFromAPI(manufacturer *nb.Manufacturer) (manufacturerItemModel, error) {
	if manufacturer == nil || manufacturer.Id == nil || *manufacturer.Id == "" {
		return manufacturerItemModel{}, fmt.Errorf("manufacturer returned no id")
	}
	return manufacturerItemModel{
		ID:          types.StringValue(*manufacturer.Id),
		Display:     types.StringValue(manufacturer.Display),
		URL:         types.StringValue(manufacturer.Url),
		NaturalSlug: types.StringValue(manufacturer.NaturalSlug),
		Name:        types.StringValue(manufacturer.Name),
		Description: types.StringValue(shared.DerefString(manufacturer.Description)),
		Created:     shared.NullableTimeValue(manufacturer.Created),
		LastUpdated: shared.NullableTimeValue(manufacturer.LastUpdated),
		NotesURL:    types.StringValue(manufacturer.NotesUrl),
	}, nil
}

func manufacturerDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "Manufacturer UUID.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "Manufacturer name.", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human-friendly display value for the manufacturer.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "API URL of the manufacturer.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the manufacturer.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "Manufacturer description.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "Manufacturer creation date (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "Manufacturer last update date (RFC3339).", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the manufacturer.", Computed: true},
	}
	if selectable {
		attributes["id"] = dsschema.StringAttribute{Description: "Manufacturer UUID. Provide either `id` or `name`.", Optional: true, Computed: true}
		attributes["name"] = dsschema.StringAttribute{Description: "Exact manufacturer name. Provide either `id` or `name`.", Optional: true, Computed: true}
	}
	return attributes
}
