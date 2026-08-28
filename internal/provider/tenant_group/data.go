package tenant_group

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type tenantGroupItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ParentID    types.String `tfsdk:"parent_id"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func tenantGroupDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "Tenant group's UUID.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "Tenant group's name.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "Tenant group's description.", Computed: true},
		"parent_id":    dsschema.StringAttribute{Description: "UUID of the parent tenant group.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "Tenant group's creation date (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "Tenant group's last update date (RFC3339).", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human friendly display value for the tenant group.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "URL of the tenant group.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the tenant group.", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the tenant group.", Computed: true},
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

func tenantGroupModelFromAPI(group *nb.TenantGroup) (tenantGroupItemModel, error) {
	if group == nil || group.Id == nil || *group.Id == "" {
		return tenantGroupItemModel{}, fmt.Errorf("tenant group returned no id")
	}
	return tenantGroupItemModel{
		ID:          types.StringValue(*group.Id),
		Name:        types.StringValue(group.Name),
		Description: types.StringValue(shared.DerefString(group.Description)),
		ParentID:    shared.NullableReferenceID(group.Parent),
		Created:     shared.NullableTimeValue(group.Created),
		LastUpdated: shared.NullableTimeValue(group.LastUpdated),
		Display:     types.StringValue(group.Display),
		URL:         types.StringValue(group.Url),
		NaturalSlug: types.StringValue(group.NaturalSlug),
		NotesURL:    types.StringValue(group.NotesUrl),
	}, nil
}
