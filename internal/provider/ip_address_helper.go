package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

func makeNATInside(id string) nb.NullableNATInside {
	var fk nb.NullableNATInside
	if id == "" {
		fk.Set(nil)
		return fk
	}
	fk.Set(&nb.NATInside{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(id),
		},
	})
	return fk
}

func nullableNATInsideStr(n nb.NullableNATInside) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

func makeIPParent(id string) nb.NullableBulkWritableIPAddressRequestParent {
	var fk nb.NullableBulkWritableIPAddressRequestParent
	if id == "" {
		fk.Set(nil)
		return fk
	}
	fk.Set(&nb.BulkWritableIPAddressRequestParent{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(id),
		},
	})
	return fk
}

func nullableIPParentStr(n nb.NullableBulkWritableIPAddressRequestParent) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

func makeIPNamespace(id string) *nb.BulkWritableIPAddressRequestNamespace {
	if id == "" {
		return nil
	}
	return &nb.BulkWritableIPAddressRequestNamespace{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(id),
		},
	}
}

func ipAddressTypeFromString(v string) (*nb.IPAddressTypeChoices, error) {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		def := nb.IPADDRESSTYPECHOICES_HOST
		return &def, nil
	}
	return nb.NewIPAddressTypeChoicesFromValue(s)
}

func ipAddressTypeToString(t *nb.IPAddressTypeChoices) types.String {
	if t == nil || *t == "" {
		return types.StringValue(string(nb.IPADDRESSTYPECHOICES_HOST))
	}
	return types.StringValue(string(*t))
}

func expandTagRefs(ctx context.Context, set types.Set) ([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, diag.Diagnostics) {
	var diags diag.Diagnostics
	if set.IsNull() || set.IsUnknown() {
		return []nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{}, diags
	}

	var tagIDs []string
	diags.Append(set.ElementsAs(ctx, &tagIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	sort.Strings(tagIDs)

	tags := make([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, 0, len(tagIDs))
	for _, t := range tagIDs {
		if t == "" {
			continue
		}
		tags = append(tags, nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(t),
			},
		})
	}
	return tags, diags
}

func flattenTagIDs(tags []nb.ApprovalWorkflowStageResponseApprovalWorkflowStage) types.Set {
	if len(tags) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, 0, len(tags))
	for _, t := range tags {
		if t.Id != nil && t.Id.String != nil && *t.Id.String != "" {
			vals = append(vals, types.StringValue(*t.Id.String))
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

func getPrefixNamespaceID(ctx context.Context, c *nb.APIClient, prefixID string) (string, error) {
	if prefixID == "" {
		return "", fmt.Errorf("prefix ID is empty")
	}
	p, httpResp, err := c.IpamAPI.IpamPrefixesRetrieve(ctx, prefixID).Execute()
	if isNotFoundResponse(httpResp) {
		return "", fmt.Errorf("parent prefix %s not found", prefixID)
	}
	if err != nil {
		return "", fmt.Errorf("%s", httpErr(err, httpResp))
	}
	if p != nil && p.Namespace != nil && p.Namespace.Id != nil && p.Namespace.Id.String != nil && *p.Namespace.Id.String != "" {
		return *p.Namespace.Id.String, nil
	}
	return "", fmt.Errorf("parent prefix %s returned no namespace id", prefixID)
}

func resolveIPAddressNamespaceID(ctx context.Context, c *nb.APIClient, ip *nb.IPAddress) (string, error) {
	if ip == nil || !ip.Parent.IsSet() || ip.Parent.Get() == nil {
		return "", fmt.Errorf("IP address has no parent prefix; cannot determine namespace_id")
	}
	parent := ip.Parent.Get()
	if nsID := extractNamespaceIDFromParentProps(parent); nsID != "" {
		return nsID, nil
	}
	parentID := nullableIPParentStr(ip.Parent).ValueString()
	if parentID == "" {
		return "", fmt.Errorf("IP address parent prefix has no id; cannot determine namespace_id")
	}
	return getPrefixNamespaceID(ctx, c, parentID)
}

// ipLocationStatePlanModifier preserves the prior state value for computed IP
// placement attributes (parent_id, namespace_id) when omitted from config,
// unless address, parent_id, or namespace_id is being changed.
type ipLocationStatePlanModifier struct{}

func useIPStateUnlessRelocated() planmodifier.String {
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
