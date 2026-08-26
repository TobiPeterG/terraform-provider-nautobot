package virtual_machine_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	vmInterfaceResourceName = "nautobot_vm_interface.test"
)

func testAccVMInterfaceConfigMinimal(name string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
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
`, name, name, name, status, name, testutil.Status)
}

func testAccVMInterfaceConfigFull(name string, vid int, cidr string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[5]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[5]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
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
  status     = "%[4]s"
}

resource "nautobot_vm_interface" "test" {
  name               = "%[1]s-if0"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:FF"
  enabled          = false
  mtu              = 1500
  description      = "created by terraform acceptance test"
  mode             = "Access"
  untagged_vlan_id = nautobot_vlan.v.id
  ip_addresses     = [nautobot_available_ip_address.ip1.id]
}
`, name, vid, cidr, status, testutil.Status)
}

func testAccVMInterfaceConfigUpdated(name string, vid int, cidr string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[5]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[5]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
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
  status     = "%[4]s"
}

resource "nautobot_vm_interface" "test" {
  name               = "%[1]s-if0-updated"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:11"
  enabled          = true
  mtu              = 9000
  description      = "updated by terraform acceptance test"
  mode             = "Tagged"
  untagged_vlan_id = nautobot_vlan.v.id
  ip_addresses     = [nautobot_available_ip_address.ip1.id]
}
`, name, vid, cidr, status, testutil.Status)
}

func testAccVMInterfaceConfigParallel(name string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
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
		name, name, name, status,
		name, testutil.Status,
		name, testutil.Status,
		name, testutil.Status,
	)
}

func TestAccVMInterfaceResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testutil.Status),
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
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_full(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-if-full-%d", seed)
	vid := testutil.AccVLANVID(seed, 20)
	cidr := testutil.AccPrefixCIDR(seed, 16)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testutil.Status),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "virtual_machine_id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testutil.CheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_update(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-if-upd-%d", seed)
	vid := testutil.AccVLANVID(seed, 21)
	cidr := testutil.AccPrefixCIDR(seed, 17)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testutil.CheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),
				),
			},
			{
				Config:             testAccVMInterfaceConfigUpdated(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVMInterfaceConfigUpdated(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0-updated"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:11"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "9000"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "updated by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Tagged"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testutil.CheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-if-drift-%d", seed)
	config := testAccVMInterfaceConfigFull(name, testutil.AccVLANVID(seed, 71), testutil.AccPrefixCIDR(seed, 71))
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(vmInterfaceResourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "virtualization/interfaces", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVMInterfaceResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-import-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
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
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-del-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vm-interface-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccVMInterfaceConfigMinimal(name), Check: testutil.DeleteResourceOutOfBand(vmInterfaceResourceName, "virtualization/interfaces")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVMInterfaceResource_parallel(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-par-%d", testutil.AccSeedForTest(t))

	resourceName1 := "nautobot_vm_interface.if1"
	resourceName2 := "nautobot_vm_interface.if2"
	resourceName3 := "nautobot_vm_interface.if3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigParallel(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", name+"-if1"),
					resource.TestCheckResourceAttr(resourceName1, "status", testutil.Status),
					resource.TestCheckResourceAttrSet(resourceName1, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName1, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName1, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName2, "name", name+"-if2"),
					resource.TestCheckResourceAttr(resourceName2, "status", testutil.Status),
					resource.TestCheckResourceAttrSet(resourceName2, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName2, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName2, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName3, "name", name+"-if3"),
					resource.TestCheckResourceAttr(resourceName3, "status", testutil.Status),
					resource.TestCheckResourceAttrSet(resourceName3, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName3, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName3, "ip_addresses.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
