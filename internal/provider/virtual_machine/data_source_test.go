package virtual_machine_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	vmDataSourceName = "data.nautobot_virtual_machine.test"
)

func TestAccVirtualMachineDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-missing-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "missing_vm" {
  name = "%[1]s-type"
}

resource "nautobot_cluster" "missing_vm" {
  name            = "%[1]s-cluster"
  cluster_type_id = nautobot_cluster_type.missing_vm.id
}

data "nautobot_virtual_machine" "test" {
  name       = %[1]q
  cluster_id = nautobot_cluster.missing_vm.id
}
`, name),
			ExpectError: regexp.MustCompile(`Virtual machine lookup failed`),
		}},
	})
}

func testAccVirtualMachineDataSourceConfigMinimal(name string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name       = "%[1]s"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

data "nautobot_virtual_machine" "test" {
  depends_on = [nautobot_virtual_machine.test]
  name       = nautobot_virtual_machine.test.name
  cluster_id = nautobot_cluster.cl.id
}
`, name, status)
}

func testAccVirtualMachineDataSourceConfigFull(name string) string {
	status := "Active"

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_tenant" "test" {
  name = "%[1]s-tenant"
}

resource "nautobot_virtual_machine" "test" {
  name                = "%[1]s"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%[2]s"

  vcpus               = 4
  memory              = 8192
  disk                = 100
  comments            = "created by terraform acceptance test"
  tenant_id           = nautobot_tenant.test.id
}

data "nautobot_virtual_machine" "test" {
  depends_on = [nautobot_virtual_machine.test]
  id         = nautobot_virtual_machine.test.id
}
`,
		name,
		status,
	)
}

func testAccVirtualMachineDataSourceTenantQualifierConfig(name string, includeTenantQualifier bool) string {
	tenantQualifier := ""
	if includeTenantQualifier {
		tenantQualifier = "tenant_id = nautobot_tenant.second.id"
	}

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_tenant" "first" {
  name = "%[1]s-first-tenant"
}

resource "nautobot_tenant" "second" {
  name = "%[1]s-second-tenant"
}

resource "nautobot_virtual_machine" "first" {
  name       = "%[1]s"
  cluster_id = nautobot_cluster.cl.id
  tenant_id  = nautobot_tenant.first.id
  status     = "%[2]s"
}

resource "nautobot_virtual_machine" "second" {
  name       = "%[1]s"
  cluster_id = nautobot_cluster.cl.id
  tenant_id  = nautobot_tenant.second.id
  status     = "%[2]s"
}

data "nautobot_virtual_machine" "test" {
  depends_on = [nautobot_virtual_machine.first, nautobot_virtual_machine.second]
  name       = "%[1]s"
  cluster_id = nautobot_cluster.cl.id
  %[3]s
}
`, name, testutil.Status, tenantQualifier)
}

func TestAccVirtualMachineDataSource_multipleMatchesWithoutTenantQualifier(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-multiple-%d", testutil.AccSeedForTest(t))
	config := testAccVirtualMachineDataSourceTenantQualifierConfig(name, false)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`Virtual machine lookup failed`),
		}},
	})
}

func TestAccVirtualMachineDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-minimal-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "cluster_id",
						"nautobot_cluster.cl", "id",
					),
					resource.TestCheckResourceAttr(vmDataSourceName, "status", "Active"),

					resource.TestCheckResourceAttr(vmDataSourceName, "vcpus", "0"),
					resource.TestCheckResourceAttr(vmDataSourceName, "memory", "0"),
					resource.TestCheckResourceAttr(vmDataSourceName, "disk", "0"),

					resource.TestCheckResourceAttr(vmDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "platform_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "software_version_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip4_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip6_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "notes_url"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "id",
						vmResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "cluster_id",
						"nautobot_cluster.cl", "id",
					),

					resource.TestCheckResourceAttr(vmDataSourceName, "status", "Active"),

					resource.TestCheckResourceAttr(vmDataSourceName, "vcpus", "4"),
					resource.TestCheckResourceAttr(vmDataSourceName, "memory", "8192"),
					resource.TestCheckResourceAttr(vmDataSourceName, "disk", "100"),

					resource.TestCheckResourceAttr(vmDataSourceName, "comments", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(vmDataSourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttr(vmDataSourceName, "platform_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "software_version_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip4_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip6_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_tenantQualifier(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-tenant-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineDataSourceTenantQualifierConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(vmDataSourceName, "id", "nautobot_virtual_machine.second", "id"),
					resource.TestCheckResourceAttrPair(vmDataSourceName, "cluster_id", "nautobot_cluster.cl", "id"),
					resource.TestCheckResourceAttrPair(vmDataSourceName, "tenant_id", "nautobot_tenant.second", "id"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
