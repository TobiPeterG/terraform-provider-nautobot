package virtual_machine_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/virtual_machine"
)

func TestVirtualMachineSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, virtual_machine.NewVirtualMachineDataSource(), "name", "cluster_id", "tenant_id")
}
