package virtual_machine_test

import (
	"fmt"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func testAccVMPrimaryIPConfigDualStack(name, ipv4CIDR, ipv6CIDR string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p4" {
  prefix = %[2]q
  status = %[4]q
}

resource "nautobot_prefix" "p6" {
  prefix = %[3]q
  status = %[4]q
}

resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%[1]s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = %[4]q
}

resource "nautobot_available_ip_address" "ip4" {
  prefix_id = nautobot_prefix.p4.id
  status    = %[4]q
}

resource "nautobot_available_ip_address" "ip6" {
  prefix_id = nautobot_prefix.p6.id
  status    = %[4]q
}

resource "nautobot_vm_interface" "if0" {
  name               = "%[1]s-if0"
  status             = %[4]q
  virtual_machine_id = nautobot_virtual_machine.vm.id
  ip_addresses = [
    nautobot_available_ip_address.ip4.id,
    nautobot_available_ip_address.ip6.id,
  ]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4.id
  primary_ip6_id     = nautobot_available_ip_address.ip6.id

  depends_on = [nautobot_vm_interface.if0]
}
`, name, ipv4CIDR, ipv6CIDR, testutil.Status)
}
