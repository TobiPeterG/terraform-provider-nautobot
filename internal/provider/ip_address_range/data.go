package ip_address_range

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

// ipAddressRangeItemModel is shared by the singular and collection data
// sources.
type ipAddressRangeItemModel struct {
	ID              types.String `tfsdk:"id"`
	StartAddress    types.String `tfsdk:"start_address"`
	EndAddress      types.String `tfsdk:"end_address"`
	NamespaceID     types.String `tfsdk:"namespace_id"`
	ParentID        types.String `tfsdk:"parent_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	CountAsUtilized types.Bool   `tfsdk:"count_as_utilized"`
	IsExclusive     types.Bool   `tfsdk:"is_exclusive"`
	Status          types.String `tfsdk:"status"`
	RoleID          types.String `tfsdk:"role_id"`
	TenantID        types.String `tfsdk:"tenant_id"`
	TagsIDs         types.List   `tfsdk:"tags_ids"`
	StartHost       types.String `tfsdk:"start_host"`
	EndHost         types.String `tfsdk:"end_host"`
	Size            types.Int64  `tfsdk:"size"`
	IPVersion       types.Int64  `tfsdk:"ip_version"`
	Created         types.String `tfsdk:"created"`
	LastUpdated     types.String `tfsdk:"last_updated"`
	Display         types.String `tfsdk:"display"`
	URL             types.String `tfsdk:"url"`
	NaturalSlug     types.String `tfsdk:"natural_slug"`
	NotesURL        types.String `tfsdk:"notes_url"`
}

// ipAddressRangeModelFromAPI maps fields that are present on the range itself.
// Callers resolve the status name and namespace separately because collection
// reads cache those API lookups.
func ipAddressRangeModelFromAPI(out *nb.IPAddressRange) (ipAddressRangeItemModel, error) {
	if out == nil {
		return ipAddressRangeItemModel{}, fmt.Errorf("IP address range response is nil")
	}
	if out.Id == nil || *out.Id == "" {
		return ipAddressRangeItemModel{}, fmt.Errorf("IP address range response has no id")
	}

	model := ipAddressRangeItemModel{
		ID:              types.StringValue(*out.Id),
		StartAddress:    types.StringValue(out.StartAddress),
		EndAddress:      types.StringValue(out.EndAddress),
		Name:            types.StringValue(shared.DerefString(out.Name)),
		Description:     types.StringValue(shared.DerefString(out.Description)),
		CountAsUtilized: types.BoolValue(out.CountAsUtilized != nil && *out.CountAsUtilized),
		IsExclusive:     types.BoolValue(out.IsExclusive != nil && *out.IsExclusive),
		RoleID:          shared.NullableReferenceID(out.Role),
		TenantID:        shared.NullableReferenceID(out.Tenant),
		TagsIDs:         shared.ReferenceIDs(out.Tags),
		StartHost:       types.StringValue(out.StartHost),
		EndHost:         types.StringValue(out.EndHost),
		Size:            types.Int64Value(int64(out.Size)),
		IPVersion:       types.Int64Value(int64(out.IpVersion)),
		Created:         shared.NullableTimeValue(out.Created),
		LastUpdated:     shared.NullableTimeValue(out.LastUpdated),
		Display:         types.StringValue(out.Display),
		URL:             types.StringValue(out.Url),
		NaturalSlug:     types.StringValue(out.NaturalSlug),
		NotesURL:        types.StringValue(out.NotesUrl),
		Status:          types.StringValue(""),
		NamespaceID:     types.StringValue(""),
	}
	if out.Parent != nil && out.Parent.Id != nil && out.Parent.Id.String != nil {
		model.ParentID = types.StringValue(*out.Parent.Id.String)
	} else {
		model.ParentID = types.StringValue("")
	}
	return model, nil
}
