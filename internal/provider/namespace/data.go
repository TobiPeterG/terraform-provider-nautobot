package namespace

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type namespaceItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func namespaceModelFromAPI(namespace *nb.Namespace) (namespaceItemModel, error) {
	if namespace == nil || namespace.Id == nil || *namespace.Id == "" {
		return namespaceItemModel{}, fmt.Errorf("namespace returned no id")
	}
	return namespaceItemModel{
		ID:          types.StringValue(*namespace.Id),
		Name:        types.StringValue(namespace.Name),
		Description: types.StringValue(shared.DerefString(namespace.Description)),
		LocationID:  shared.NullableReferenceID(namespace.Location),
		TenantID:    shared.NullableReferenceID(namespace.Tenant),
		Created:     shared.NullableTimeValue(namespace.Created),
		LastUpdated: shared.NullableTimeValue(namespace.LastUpdated),
		Display:     types.StringValue(namespace.Display),
		URL:         types.StringValue(namespace.Url),
		NaturalSlug: types.StringValue(namespace.NaturalSlug),
		NotesURL:    types.StringValue(namespace.NotesUrl),
	}, nil
}

func namespaceDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "Namespace UUID.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "Namespace name.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "Namespace description.", Computed: true},
		"location_id":  dsschema.StringAttribute{Description: "UUID of the location associated with the namespace.", Computed: true},
		"tenant_id":    dsschema.StringAttribute{Description: "UUID of the tenant associated with the namespace.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "Namespace creation date (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "Namespace last update date (RFC3339).", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human-friendly display value for the namespace.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "API URL of the namespace.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the namespace.", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the namespace.", Computed: true},
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
