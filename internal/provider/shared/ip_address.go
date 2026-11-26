package shared

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

func NullableNATInside(id string) nb.NullableNATInside {
	var fk nb.NullableNATInside
	if id == "" {
		fk.Set(nil)
		return fk
	}
	fk.Set(&nb.NATInside{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: StringPtr(id),
		},
	})
	return fk
}

func NullableNATInsideID(n nb.NullableNATInside) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

func NullableIPParent(id string) nb.NullableBulkWritableIPAddressRequestParent {
	var fk nb.NullableBulkWritableIPAddressRequestParent
	if id == "" {
		fk.Set(nil)
		return fk
	}
	fk.Set(&nb.BulkWritableIPAddressRequestParent{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: StringPtr(id),
		},
	})
	return fk
}

func NullableIPParentID(n nb.NullableBulkWritableIPAddressRequestParent) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

func IPNamespaceReference(id string) *nb.BulkWritableIPAddressRangeRequestNamespace {
	if id == "" {
		return nil
	}
	return &nb.BulkWritableIPAddressRangeRequestNamespace{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: StringPtr(id),
		},
	}
}

func IPAddressTypeFromString(v string) (*nb.IPAddressTypeChoices, error) {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		def := nb.IPADDRESSTYPECHOICES_HOST
		return &def, nil
	}
	return nb.NewIPAddressTypeChoicesFromValue(s)
}

func IPAddressTypeValue(t *nb.IPAddressTypeChoices) types.String {
	if t == nil || *t == "" {
		return types.StringValue(string(nb.IPADDRESSTYPECHOICES_HOST))
	}
	return types.StringValue(string(*t))
}

func ExpandTagSetRefs(ctx context.Context, set types.Set) ([]nb.BulkWritableCableRequestStatus, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return []nb.BulkWritableCableRequestStatus{}, diags
	}

	var tagIDs []string
	diags.Append(set.ElementsAs(ctx, &tagIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	sort.Strings(tagIDs)
	return APIReferences(tagIDs), diags
}

func ReferenceIDSet(references []nb.BulkWritableCableRequestStatus) types.Set {
	if len(references) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, 0, len(references))
	for _, ref := range references {
		if ref.Id != nil && ref.Id.String != nil && *ref.Id.String != "" {
			vals = append(vals, types.StringValue(*ref.Id.String))
		}
	}
	return types.SetValueMust(types.StringType, vals)
}

func extractNamespaceIDFromParentProps(parent *nb.BulkWritableIPAddressRequestParent) string {
	if parent == nil || len(parent.AdditionalProperties) == 0 {
		return ""
	}
	nsRaw, ok := parent.AdditionalProperties["namespace"]
	if !ok || nsRaw == nil {
		return ""
	}
	nsMap, ok := nsRaw.(map[string]any)
	if !ok {
		return ""
	}
	idStr, ok := nsMap["id"].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(idStr)
}

func GetPrefixNamespaceID(ctx context.Context, c *nb.APIClient, prefixID string) (string, error) {
	if prefixID == "" {
		return "", fmt.Errorf("prefix ID is empty")
	}
	p, httpResp, err := c.IpamAPI.IpamPrefixesRetrieve(ctx, prefixID).Execute()
	if IsNotFoundResponse(httpResp) {
		return "", fmt.Errorf("parent prefix %s not found", prefixID)
	}
	if err != nil {
		return "", fmt.Errorf("%s", HTTPError(err, httpResp))
	}
	if p != nil && p.Namespace != nil && p.Namespace.Id != nil && p.Namespace.Id.String != nil && *p.Namespace.Id.String != "" {
		return *p.Namespace.Id.String, nil
	}
	return "", fmt.Errorf("parent prefix %s returned no namespace id", prefixID)
}

func ResolveIPAddressNamespaceID(ctx context.Context, c *nb.APIClient, ip *nb.IPAddress) (string, error) {
	if ip == nil || !ip.Parent.IsSet() || ip.Parent.Get() == nil {
		return "", fmt.Errorf("IP address has no parent prefix; cannot determine namespace_id")
	}
	parent := ip.Parent.Get()
	if nsID := extractNamespaceIDFromParentProps(parent); nsID != "" {
		return nsID, nil
	}
	parentID := NullableIPParentID(ip.Parent).ValueString()
	if parentID == "" {
		return "", fmt.Errorf("IP address parent prefix has no id; cannot determine namespace_id")
	}
	return GetPrefixNamespaceID(ctx, c, parentID)
}

// ipLocationStatePlanModifier preserves the prior state value for computed IP
// placement attributes (parent_id, namespace_id) when omitted from config,
// unless address, parent_id, or namespace_id is being changed.
type ipLocationStatePlanModifier struct{}

func UseIPStateUnlessRelocated() planmodifier.String {
	return ipLocationStatePlanModifier{}
}

func (m ipLocationStatePlanModifier) Description(_ context.Context) string {
	return "Preserves the state value unless address, parent_id, or namespace_id changes."
}

func (m ipLocationStatePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ipLocationStatePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}

	var planAddr, stateAddr types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("address"), &planAddr)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("address"), &stateAddr)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !planAddr.Equal(stateAddr) {
		return
	}

	var cfgParent, stateParent types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("parent_id"), &cfgParent)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("parent_id"), &stateParent)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfgParent.IsNull() && !cfgParent.Equal(stateParent) {
		return
	}

	var cfgNS, stateNS types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("namespace_id"), &cfgNS)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("namespace_id"), &stateNS)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !cfgNS.IsNull() && !cfgNS.Equal(stateNS) {
		return
	}

	resp.PlanValue = req.StateValue
}

// GetNamespaceIDByName resolves a Nautobot IPAM namespace UUID by its exact name.
func GetNamespaceIDByName(ctx context.Context, client *nb.APIClient, name string) (string, error) {
	rsp, httpResp, err := client.IpamAPI.
		IpamNamespacesList(ctx).
		Name([]string{name}).
		Execute()
	if err != nil {
		return "", fmt.Errorf("lookup namespace %q: %s", name, HTTPError(err, httpResp))
	}
	if rsp == nil || len(rsp.Results) == 0 {
		return "", fmt.Errorf("namespace %q not found", name)
	}
	if len(rsp.Results) > 1 {
		return "", fmt.Errorf("multiple namespaces matched %q", name)
	}
	if rsp.Results[0].Id == nil || *rsp.Results[0].Id == "" {
		return "", fmt.Errorf("namespace %q returned no id", name)
	}
	return *rsp.Results[0].Id, nil
}

// RequiresReplaceIfIPHostChanged marks an IP address resource for replacement
// when the host component of the CIDR changes, because Nautobot rejects
// in-place host changes ("Host address cannot be changed once created") while
// still allowing mask-length updates in place.
func RequiresReplaceIfIPHostChanged() planmodifier.String {
	return requiresReplaceIfIPHostChangedModifier{}
}

type requiresReplaceIfIPHostChangedModifier struct{}

func (m requiresReplaceIfIPHostChangedModifier) Description(_ context.Context) string {
	return "Requires replacement if the host portion of the IP address changes."
}

func (m requiresReplaceIfIPHostChangedModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m requiresReplaceIfIPHostChangedModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.Equal(req.PlanValue) {
		return
	}

	statePrefix, stateErr := netip.ParsePrefix(strings.TrimSpace(req.StateValue.ValueString()))
	planPrefix, planErr := netip.ParsePrefix(strings.TrimSpace(req.PlanValue.ValueString()))
	if stateErr != nil || planErr != nil {
		resp.RequiresReplace = true
		return
	}
	if statePrefix.Addr() != planPrefix.Addr() {
		resp.RequiresReplace = true
	}
}
