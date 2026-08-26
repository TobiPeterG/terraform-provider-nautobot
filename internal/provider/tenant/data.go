package tenant

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type tenantItemModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Comments      types.String `tfsdk:"comments"`
	TenantGroupID types.String `tfsdk:"tenant_group_id"`
	Created       types.String `tfsdk:"created"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

func tenantModelFromAPI(tenant *nb.Tenant) (tenantItemModel, error) {
	if tenant == nil || tenant.Id == nil || *tenant.Id == "" {
		return tenantItemModel{}, fmt.Errorf("tenant returned no id")
	}
	return tenantItemModel{
		ID:            types.StringValue(*tenant.Id),
		Name:          types.StringValue(tenant.Name),
		Description:   types.StringValue(shared.DerefString(tenant.Description)),
		Comments:      types.StringValue(shared.DerefString(tenant.Comments)),
		TenantGroupID: shared.NullableReferenceID(tenant.TenantGroup),
		Created:       shared.NullableTimeValue(tenant.Created),
		LastUpdated:   shared.NullableTimeValue(tenant.LastUpdated),
		Display:       types.StringValue(tenant.Display),
		URL:           types.StringValue(tenant.Url),
		NaturalSlug:   types.StringValue(tenant.NaturalSlug),
		NotesURL:      types.StringValue(tenant.NotesUrl),
	}, nil
}

func tenantDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"name":            dsschema.StringAttribute{Description: "The name of the tenant to retrieve.", Optional: true, Computed: true},
		"id":              dsschema.StringAttribute{Description: "Tenant's UUID. Provide either `id` or `name`.", Optional: true, Computed: true},
		"description":     dsschema.StringAttribute{Description: "Tenant's description.", Computed: true},
		"comments":        dsschema.StringAttribute{Description: "Tenant's comments.", Computed: true},
		"tenant_group_id": dsschema.StringAttribute{Description: "UUID of the tenant group this tenant belongs to.", Computed: true},
		"created":         dsschema.StringAttribute{Description: "Tenant's creation date (RFC3339).", Computed: true},
		"last_updated":    dsschema.StringAttribute{Description: "Tenant's last update date (RFC3339).", Computed: true},
		"display":         dsschema.StringAttribute{Description: "Human friendly display value for the tenant.", Computed: true},
		"url":             dsschema.StringAttribute{Description: "URL of the tenant.", Computed: true},
		"natural_slug":    dsschema.StringAttribute{Description: "Natural slug for the tenant.", Computed: true},
		"notes_url":       dsschema.StringAttribute{Description: "Notes URL for the tenant.", Computed: true},
	}
	if !selectable {
		for _, name := range []string{"id", "name"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = false
			attribute.Computed = true
			attributes[name] = attribute
		}
	}
	return attributes
}
