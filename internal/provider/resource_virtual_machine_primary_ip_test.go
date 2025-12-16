package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmPrimaryIPResourceName = "nautobot_vm_primary_ip.test"
)

func testAccVMPrimaryIPConfigIPv4Single(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%s"
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_available_ip_address" "ip4" {
  prefix_id = "%s"
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [nautobot_available_ip_address.ip4.id]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4.id

  depends_on = [nautobot_vm_interface.if0]
}
`,
		name, name, testTenantID,
		name, status,
		testPrefixID,
		name, testStatus,
	)
}

func testAccVMPrimaryIPConfigIPv4UpdateA(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%s"
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_available_ip_address" "ip4a" {
  prefix_id = "%s"
  status    = "Active"
}

resource "nautobot_available_ip_address" "ip4b" {
  prefix_id = "%s"
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [
    nautobot_available_ip_address.ip4a.id,
    nautobot_available_ip_address.ip4b.id,
  ]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4a.id

  depends_on = [nautobot_vm_interface.if0]
}
`,
		name, name, testTenantID,
		name, status,
		testPrefixID,
		testPrefixID,
		name, testStatus,
	)
}

func testAccVMPrimaryIPConfigIPv4UpdateB(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%s"
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_available_ip_address" "ip4a" {
  prefix_id = "%s"
  status    = "Active"
}

resource "nautobot_available_ip_address" "ip4b" {
  prefix_id = "%s"
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [
    nautobot_available_ip_address.ip4a.id,
    nautobot_available_ip_address.ip4b.id,
  ]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4b.id

  depends_on = [nautobot_vm_interface.if0]
}
`,
		name, name, testTenantID,
		name, status,
		testPrefixID,
		testPrefixID,
		name, testStatus,
	)
}

func TestAccVMPrimaryIPResource_ipv4Single(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-primary-ip-v4single-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_updateIP(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-primary-ip-update-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4UpdateA(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4a", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config:             testAccVMPrimaryIPConfigIPv4UpdateB(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVMPrimaryIPConfigIPv4UpdateB(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4b", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-primary-ip-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name),
			},
			{
				ResourceName:      vmPrimaryIPResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-primary-ip-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
