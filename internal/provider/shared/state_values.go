package shared

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

func NullableTimeValue(value nb.NullableTime) types.String {
	if value.IsSet() && value.Get() != nil {
		return types.StringValue(value.Get().Format(time.RFC3339Nano))
	}
	return types.StringNull()
}

func NullableRFC3339InstantValue(value nb.NullableTime) RFC3339Instant {
	if value.IsSet() && value.Get() != nil {
		return NewRFC3339InstantValue(*value.Get())
	}
	return NewRFC3339InstantNull()
}

func NullableReferenceID(reference nb.NullableApprovalWorkflowUser) types.String {
	if reference.IsSet() {
		if value := reference.Get(); value != nil && value.Id != nil && value.Id.String != nil {
			return types.StringValue(*value.Id.String)
		}
	}
	return types.StringValue("")
}

func NullableSoftwareVersionID(reference nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion) types.String {
	if reference.IsSet() {
		if value := reference.Get(); value != nil && value.Id != nil && value.Id.String != nil {
			return types.StringValue(*value.Id.String)
		}
	}
	return types.StringValue("")
}
