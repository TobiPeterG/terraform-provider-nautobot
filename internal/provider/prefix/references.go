package prefix

import (
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

func prefixParentReference(id string) nb.NullableBulkWritablePrefixRequestParent {
	var reference nb.NullableBulkWritablePrefixRequestParent
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.BulkWritablePrefixRequestParent{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)},
	})
	return reference
}

func prefixRIRReference(id string) nb.NullableBulkWritablePrefixRequestRir {
	var reference nb.NullableBulkWritablePrefixRequestRir
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.BulkWritablePrefixRequestRir{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)},
	})
	return reference
}
