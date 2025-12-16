package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmInterfaceResourceName  = "nautobot_vm_interface.test"
	testVMInterfaceStatus    = "Active"
	testVMInterfaceIPAddress = "e81bc81b-0db0-5bc4-af61-c0e8c5020987"
	testVMInterfaceVLAN      = "9feba4b3-9fc8-5298-a6c1-8f0f77378f21"
)

func testAccVMInterfaceConfigMinimal(name string) string {
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

resource "nautobot_vm_interface" "test" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}
`, name, name, testTenantID, name, status, name, testVMInterfaceStatus)
}

func testAccVMInterfaceConfigFull(name string) string {
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

resource "nautobot_vm_interface" "test" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:FF"
  enabled          = false
  mtu              = 1500
  description      = "created by terraform acceptance test"
  mode             = "Access"
  untagged_vlan_id = "%s"
  ip_addresses     = ["%s"]
}
`, name, name, testTenantID, name, status, name, testVMInterfaceStatus, testVMInterfaceVLAN, testVMInterfaceIPAddress)
}

func testAccVMInterfaceConfigUpdated(name string) string {
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

resource "nautobot_vm_interface" "test" {
  name               = "%s-if0-updated"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:11"
  enabled          = true
  mtu              = 9000
  description      = "updated by terraform acceptance test"
  mode             = "Tagged"
  untagged_vlan_id = "%s"
  ip_addresses     = ["%s"]
}
`, name, name, testTenantID, name, status, name, testVMInterfaceStatus, testVMInterfaceVLAN, testVMInterfaceIPAddress)
}

func testAccVMInterfaceConfigParallel(name string) string {
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

resource "nautobot_vm_interface" "if1" {
  name               = "%s-if1"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}

resource "nautobot_vm_interface" "if2" {
  name               = "%s-if2"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}

resource "nautobot_vm_interface" "if3" {
  name               = "%s-if3"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}
`,
		name, name, testTenantID, name, status,
		name, testVMInterfaceStatus,
		name, testVMInterfaceStatus,
		name, testVMInterfaceStatus,
	)
}

func TestAccVMInterfaceResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "virtual_machine_id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "untagged_vlan_id", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "virtual_machine_id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "untagged_vlan_id", testVMInterfaceVLAN),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						vmInterfaceResourceName,
						"ip_addresses.*",
						testVMInterfaceIPAddress,
					),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "untagged_vlan_id", testVMInterfaceVLAN),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						vmInterfaceResourceName,
						"ip_addresses.*",
						testVMInterfaceIPAddress,
					),
				),
			},
			{
				Config:             testAccVMInterfaceConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVMInterfaceConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0-updated"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:11"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "9000"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "updated by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Tagged"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "untagged_vlan_id", testVMInterfaceVLAN),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						vmInterfaceResourceName,
						"ip_addresses.*",
						testVMInterfaceIPAddress,
					),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
			},
			{
				ResourceName:      vmInterfaceResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_parallel(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-par-%d", time.Now().Unix())

	resourceName1 := "nautobot_vm_interface.if1"
	resourceName2 := "nautobot_vm_interface.if2"
	resourceName3 := "nautobot_vm_interface.if3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigParallel(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", name+"-if1"),
					resource.TestCheckResourceAttr(resourceName1, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttrSet(resourceName1, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName1, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName1, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName2, "name", name+"-if2"),
					resource.TestCheckResourceAttr(resourceName2, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttrSet(resourceName2, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName2, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName2, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName3, "name", name+"-if3"),
					resource.TestCheckResourceAttr(resourceName3, "status", testVMInterfaceStatus),
					resource.TestCheckResourceAttrSet(resourceName3, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName3, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName3, "ip_addresses.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
