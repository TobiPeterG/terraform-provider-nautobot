package virtual_machine

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"

	nb "github.com/nautobot/go-nautobot/v3"
)

func (r *VMInterfaceResource) reconcileIPAssignments(ctx context.Context, vmInterfaceID string, desired []string) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	currentModel, found, readDiagnostics := r.readModel(ctx, vmInterfaceID)
	diagnostics.Append(readDiagnostics...)
	if diagnostics.HasError() {
		return diagnostics
	}
	if !found {
		diagnostics.AddError("Failed to read VM interface", "updated VM interface was not found")
		return diagnostics
	}

	var current []string
	diagnostics.Append(currentModel.IPAddresses.ElementsAs(ctx, &current, false)...)
	if diagnostics.HasError() {
		return diagnostics
	}

	currentSet := shared.SliceToSet(current)
	desiredSet := shared.SliceToSet(desired)
	for _, ipAddressID := range shared.SetDiff(currentSet, desiredSet) {
		if err := r.removeIPAddressFromVMInterface(ctx, ipAddressID, vmInterfaceID); err != nil {
			diagnostics.AddError("Failed to remove IP address from VM interface", err.Error())
			return diagnostics
		}
	}
	for _, ipAddressID := range shared.SetDiff(desiredSet, currentSet) {
		if err := r.assignIPAddressToVMInterface(ctx, ipAddressID, vmInterfaceID); err != nil {
			diagnostics.AddError("Failed to assign IP address to VM interface", err.Error())
			return diagnostics
		}
	}
	return diagnostics
}

func (r *VMInterfaceResource) assignIPAddressToVMInterface(ctx context.Context, ipAddressID, vmInterfaceID string) error {
	request := nb.IPAddressToInterfaceRequest{
		IpAddress:   shared.APIReference(ipAddressID),
		VmInterface: shared.NullableReference(vmInterfaceID),
	}

	_, httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressToInterfaceCreate(ctx).
		IPAddressToInterfaceRequest(request).
		Execute()
	if err != nil {
		return shared.HTTPErrorAsError(err, httpResp)
	}
	return nil
}

func (r *VMInterfaceResource) removeIPAddressFromVMInterface(ctx context.Context, ipAddressID, vmInterfaceID string) error {
	client := r.client.Client

	list, httpResp, err := client.IpamAPI.
		IpamIpAddressToInterfaceList(ctx).
		IpAddress([]string{ipAddressID}).
		VmInterface([]string{vmInterfaceID}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to list IPAddressToInterface: %s", shared.HTTPError(err, httpResp))
	}

	selector := fmt.Sprintf("IP %s on VM interface %s", ipAddressID, vmInterfaceID)
	if err := shared.ExactMatchError("IP address assignment", selector, len(list.Results)); err != nil {
		return err
	}
	if list.Results[0].Id == nil || *list.Results[0].Id == "" {
		return fmt.Errorf("IP address assignment for %s returned no id", selector)
	}

	httpResp, err = client.IpamAPI.IpamIpAddressToInterfaceDestroy(ctx, *list.Results[0].Id).Execute()
	if err != nil {
		return shared.HTTPErrorAsError(err, httpResp)
	}
	return nil
}
