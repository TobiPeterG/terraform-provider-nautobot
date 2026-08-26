package prefix

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

type prefixItemModel struct {
	ID            types.String `tfsdk:"id"`
	Prefix        types.String `tfsdk:"prefix"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	ParentID      types.String `tfsdk:"parent_id"`
	RoleID        types.String `tfsdk:"role_id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	RIRID         types.String `tfsdk:"rir_id"`
	NamespaceID   types.String `tfsdk:"namespace_id"`
	VLANID        types.String `tfsdk:"vlan_id"`
	Created       types.String `tfsdk:"created"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	Network       types.String `tfsdk:"network"`
	Broadcast     types.String `tfsdk:"broadcast"`
	PrefixLength  types.Int64  `tfsdk:"prefix_length"`
	IPVersion     types.Int64  `tfsdk:"ip_version"`
	DateAllocated types.String `tfsdk:"date_allocated"`
	TagsIDs       types.List   `tfsdk:"tags_ids"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

func prefixDataAttributes(selectable bool) map[string]dsschema.Attribute {
	attributes := map[string]dsschema.Attribute{
		"id":             dsschema.StringAttribute{Description: "The UUID of the prefix.", Computed: true},
		"prefix":         dsschema.StringAttribute{Description: "The prefix in CIDR notation.", Computed: true},
		"description":    dsschema.StringAttribute{Description: "Description of the prefix.", Computed: true},
		"status":         dsschema.StringAttribute{Description: "The status of the prefix (name).", Computed: true},
		"parent_id":      dsschema.StringAttribute{Description: "The ID of the parent of this prefix.", Computed: true},
		"role_id":        dsschema.StringAttribute{Description: "The ID of the role associated with the prefix.", Computed: true},
		"tenant_id":      dsschema.StringAttribute{Description: "The ID of the tenant associated with the prefix.", Computed: true},
		"rir_id":         dsschema.StringAttribute{Description: "The ID of the RIR associated with the prefix.", Computed: true},
		"namespace_id":   dsschema.StringAttribute{Description: "The ID of the namespace associated with the prefix.", Computed: true},
		"vlan_id":        dsschema.StringAttribute{Description: "The UUID of the VLAN the prefix belongs to.", Computed: true},
		"created":        dsschema.StringAttribute{Description: "The creation date of the prefix (RFC3339).", Computed: true},
		"last_updated":   dsschema.StringAttribute{Description: "The last update date of the prefix (RFC3339).", Computed: true},
		"network":        dsschema.StringAttribute{Description: "IPv4 or IPv6 network address.", Computed: true},
		"broadcast":      dsschema.StringAttribute{Description: "IPv4 or IPv6 broadcast address.", Computed: true},
		"prefix_length":  dsschema.Int64Attribute{Description: "Length of the network prefix, in bits.", Computed: true},
		"ip_version":     dsschema.Int64Attribute{Description: "IP version of the prefix (4 or 6).", Computed: true},
		"date_allocated": dsschema.StringAttribute{Description: "Date this prefix was allocated or reserved (RFC3339).", Computed: true},
		"tags_ids":       dsschema.ListAttribute{Description: "The IDs of the tags associated with the prefix.", Computed: true, ElementType: types.StringType},
		"display":        dsschema.StringAttribute{Description: "Human-friendly display value.", Computed: true},
		"url":            dsschema.StringAttribute{Description: "API URL of the prefix.", Computed: true},
		"natural_slug":   dsschema.StringAttribute{Description: "Natural slug for the prefix.", Computed: true},
		"notes_url":      dsschema.StringAttribute{Description: "Notes URL for the prefix.", Computed: true},
	}
	if selectable {
		for _, name := range []string{"id", "prefix", "namespace_id"} {
			attribute := attributes[name].(dsschema.StringAttribute)
			attribute.Optional = true
			attributes[name] = attribute
		}
	}
	return attributes
}

func prefixModelFromAPI(prefix *nb.Prefix, statusName func(string) (string, error)) (prefixItemModel, error) {
	if prefix == nil || prefix.Id == nil || *prefix.Id == "" {
		return prefixItemModel{}, fmt.Errorf("prefix response has no id")
	}
	var model prefixItemModel

	model.ID = types.StringValue(*prefix.Id)
	model.Prefix = types.StringValue(prefix.Prefix)
	model.Description = types.StringValue(shared.DerefString(prefix.Description))
	model.Created = shared.NullableTimeValue(prefix.Created)
	model.LastUpdated = shared.NullableTimeValue(prefix.LastUpdated)

	resolvedStatusName := ""
	if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
		if statusID := *prefix.Status.Id.String; statusID != "" {
			name, err := statusName(statusID)
			if err != nil {
				return prefixItemModel{}, err
			}
			resolvedStatusName = name
		}
	}
	model.Status = types.StringValue(resolvedStatusName)

	parentID := ""
	if prefix.Parent.IsSet() {
		if parent := prefix.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
			parentID = *parent.Id.String
		}
	}
	model.ParentID = types.StringValue(parentID)
	model.TenantID = shared.NullableReferenceID(prefix.Tenant)
	model.RoleID = shared.NullableReferenceID(prefix.Role)

	rirID := ""
	if prefix.Rir.IsSet() {
		if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
			rirID = *rir.Id.String
		}
	}
	model.RIRID = types.StringValue(rirID)

	namespaceID := ""
	if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
		namespaceID = *prefix.Namespace.Id.String
	}
	model.NamespaceID = types.StringValue(namespaceID)
	model.VLANID = shared.NullableReferenceID(prefix.Vlan)
	model.Network = types.StringValue(prefix.Network)
	model.Broadcast = types.StringValue(prefix.Broadcast)
	model.PrefixLength = types.Int64Value(int64(prefix.PrefixLength))
	model.IPVersion = types.Int64Value(int64(prefix.IpVersion))
	model.DateAllocated = shared.NullableTimeValue(prefix.DateAllocated)
	model.TagsIDs = shared.ReferenceIDs(prefix.Tags)
	model.Display = types.StringValue(prefix.Display)
	model.URL = types.StringValue(prefix.Url)
	model.NaturalSlug = types.StringValue(prefix.NaturalSlug)
	model.NotesURL = types.StringValue(prefix.NotesUrl)

	return model, nil
}
