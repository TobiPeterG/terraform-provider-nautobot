package shared

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

// APIReference converts a UUID into the SDK's generated reference type. The
// generated type names are misleading, but the wire representation is the same
// {"id": "..."} object used for statuses, tags, and several foreign keys.
func APIReference(id string) nb.BulkWritableCableRequestStatus {
	return nb.BulkWritableCableRequestStatus{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: StringPtr(id)},
	}
}

func APIReferencePointer(id string) *nb.BulkWritableCableRequestStatus {
	reference := APIReference(id)
	return &reference
}

func APIReferences(ids []string) []nb.BulkWritableCableRequestStatus {
	references := make([]nb.BulkWritableCableRequestStatus, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			references = append(references, APIReference(id))
		}
	}
	return references
}

func ReferenceIDs(references []nb.BulkWritableCableRequestStatus) types.List {
	values := make([]attr.Value, 0, len(references))
	for _, reference := range references {
		if reference.Id != nil && reference.Id.String != nil && *reference.Id.String != "" {
			values = append(values, types.StringValue(*reference.Id.String))
		}
	}
	return types.ListValueMust(types.StringType, values)
}

// NullableReference builds a nullable API reference from a UUID. An empty ID
// produces a set-but-nil value, which clears the relation on PATCH.
func NullableReference(id string) nb.NullableApprovalWorkflowUser {
	var reference nb.NullableApprovalWorkflowUser
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.ApprovalWorkflowUser{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: StringPtr(id)},
	})
	return reference
}
