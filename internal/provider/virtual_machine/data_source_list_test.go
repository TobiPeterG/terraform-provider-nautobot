package virtual_machine_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	virtualMachinesDataSourceName = "data.nautobot_virtual_machines.test"
)

func testAccVirtualMachinesDataSourceConfigBasic(base string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm1" {
  name       = "%[1]s-1"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

resource "nautobot_virtual_machine" "vm2" {
  name       = "%[1]s-2"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

resource "nautobot_virtual_machine" "vm3" {
  name       = "%[1]s-3"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

data "nautobot_virtual_machines" "test" {
  depends_on = [
    nautobot_virtual_machine.vm1,
    nautobot_virtual_machine.vm2,
    nautobot_virtual_machine.vm3,
  ]
}
`,
		base,
		status,
	)
}

func testAccVirtualMachinesDataSourceConfigFull(base string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

# minimal VM (to ensure mixed content works)
resource "nautobot_virtual_machine" "vm1" {
  name       = "%[1]s-1"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

# full VM A
resource "nautobot_virtual_machine" "vm2" {
  name                = "%[1]s-2"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%[2]s"

  vcpus               = 4
  memory              = 8192
  disk                = 100
  comments            = "vm2 created by terraform acceptance test"
}

# full VM B (different values)
resource "nautobot_virtual_machine" "vm3" {
  name                = "%[1]s-3"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%[2]s"

  vcpus               = 8
  memory              = 16384
  disk                = 200
  comments            = "vm3 created by terraform acceptance test"
}

data "nautobot_virtual_machines" "test" {
  depends_on = [
    nautobot_virtual_machine.vm1,
    nautobot_virtual_machine.vm2,
    nautobot_virtual_machine.vm3,
  ]
}
`,
		base,
		status,
	)
}

func TestAccVirtualMachinesDataSource_basic(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-vms-basic-%d", testutil.AccSeedForTest(t))

	vm1 := base + "-1"
	vm2 := base + "-2"
	vm3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachinesDataSourceConfigBasic(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(virtualMachinesDataSourceName, "virtual_machines", 3),

					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm1),
					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm2),
					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm3),

					testutil.CheckVMInListHasAttrs(virtualMachinesDataSourceName, vm1, map[string]string{
						"name":                vm1,
						"status":              "Active",
						"tenant_id":           "",
						"platform_id":         "",
						"role_id":             "",
						"software_version_id": "",
						"primary_ip4_id":      "",
						"primary_ip6_id":      "",
						"vcpus":               "0",
						"memory":              "0",
						"disk":                "0",
						"comments":            "",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachinesDataSource_full(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-vms-full-%d", testutil.AccSeedForTest(t))

	vm1 := base + "-1"
	vm2 := base + "-2"
	vm3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachinesDataSourceConfigFull(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(virtualMachinesDataSourceName, "virtual_machines", 3),

					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm1),
					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm2),
					testutil.FindListIndexByAttr(virtualMachinesDataSourceName, "virtual_machines", "name", vm3),

					testutil.CheckVMInListHasAttrs(virtualMachinesDataSourceName, vm1, map[string]string{
						"name":                vm1,
						"status":              "Active",
						"tenant_id":           "",
						"platform_id":         "",
						"role_id":             "",
						"software_version_id": "",
						"primary_ip4_id":      "",
						"primary_ip6_id":      "",
						"vcpus":               "0",
						"memory":              "0",
						"disk":                "0",
						"comments":            "",
					}),

					testutil.CheckVMInListHasAttrs(virtualMachinesDataSourceName, vm2, map[string]string{
						"name":                vm2,
						"status":              "Active",
						"tenant_id":           "",
						"platform_id":         "",
						"role_id":             "",
						"software_version_id": "",
						"primary_ip4_id":      "",
						"primary_ip6_id":      "",
						"vcpus":               "4",
						"memory":              "8192",
						"disk":                "100",
						"comments":            "vm2 created by terraform acceptance test",
					}),

					testutil.CheckVMInListHasAttrs(virtualMachinesDataSourceName, vm3, map[string]string{
						"name":                vm3,
						"status":              "Active",
						"tenant_id":           "",
						"platform_id":         "",
						"role_id":             "",
						"software_version_id": "",
						"primary_ip4_id":      "",
						"primary_ip6_id":      "",
						"vcpus":               "8",
						"memory":              "16384",
						"disk":                "200",
						"comments":            "vm3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
