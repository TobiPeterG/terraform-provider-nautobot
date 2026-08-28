package shared_test

import (
	"testing"
	"time"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestNullableTimeValue(t *testing.T) {
	var unset nb.NullableTime
	if got := shared.NullableTimeValue(unset); !got.IsNull() {
		t.Fatalf("expected unset time to become null, got %v", got)
	}
	var explicitNull nb.NullableTime
	explicitNull.Set(nil)
	if got := shared.NullableTimeValue(explicitNull); !got.IsNull() {
		t.Fatalf("expected explicit null time to become null, got %v", got)
	}
	want := time.Date(2026, time.July, 22, 12, 34, 56, 123456000, time.FixedZone("test", 2*60*60))
	var value nb.NullableTime
	value.Set(&want)
	if got := shared.NullableTimeValue(value).ValueString(); got != want.Format(time.RFC3339Nano) {
		t.Fatalf("expected %q, got %q", want.Format(time.RFC3339Nano), got)
	}
}

func TestNullableReferenceID(t *testing.T) {
	var unset nb.NullableApprovalWorkflowUser
	if got := shared.NullableReferenceID(unset).ValueString(); got != "" {
		t.Fatalf("expected unset FK to become empty, got %q", got)
	}
	var explicitNull nb.NullableApprovalWorkflowUser
	explicitNull.Set(nil)
	if got := shared.NullableReferenceID(explicitNull).ValueString(); got != "" {
		t.Fatalf("expected explicit null FK to become empty, got %q", got)
	}
	var missingID nb.NullableApprovalWorkflowUser
	missingID.Set(&nb.ApprovalWorkflowUser{})
	if got := shared.NullableReferenceID(missingID).ValueString(); got != "" {
		t.Fatalf("expected FK without an ID to become empty, got %q", got)
	}
	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	if got := shared.NullableReferenceID(shared.NullableReference(want)).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNullableSoftwareVersionID(t *testing.T) {
	var unset nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	if got := shared.NullableSoftwareVersionID(unset).ValueString(); got != "" {
		t.Fatalf("expected unset software version to become empty, got %q", got)
	}
	var missingID nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	missingID.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{})
	if got := shared.NullableSoftwareVersionID(missingID).ValueString(); got != "" {
		t.Fatalf("expected software version without an ID to become empty, got %q", got)
	}
	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	var populated nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	populated.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: &want}})
	if got := shared.NullableSoftwareVersionID(populated).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
