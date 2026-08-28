package virtual_machine

import (
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

func softwareVersionReference(id string) nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion {
	var reference nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)},
	})
	return reference
}

func primaryIPv4Reference(id string) nb.NullablePrimaryIPv4 {
	var reference nb.NullablePrimaryIPv4
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.PrimaryIPv4{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)},
	})
	return reference
}

func primaryIPv6Reference(id string) nb.NullablePrimaryIPv6 {
	var reference nb.NullablePrimaryIPv6
	if id == "" {
		reference.Set(nil)
		return reference
	}
	reference.Set(&nb.PrimaryIPv6{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: shared.StringPtr(id)},
	})
	return reference
}
