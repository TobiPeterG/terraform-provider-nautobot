package vlan_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const vlanGroupResourceName = "nautobot_vlan_group.test"

func testAccVLANGroupResourceConfig(name, description, vlanRange string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan_group" "test" {
  name        = %[1]q
  description = %[2]q
  range       = %[3]q
}
`, name, description, vlanRange)
}

func testAccVLANGroupResourceParallelConfig(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan_group" "first" {
  name  = %[1]q
  range = "1-100"
}
resource "nautobot_vlan_group" "second" {
  name  = "%[1]s-2"
  range = "101-200"
}
resource "nautobot_vlan_group" "third" {
  name  = "%[1]s-3"
  range = "201-300"
}
`, name)
}

func TestAccVLANGroupResource_minimal(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vlan-group-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVLANGroupResourceConfig(name, "", "1-4094"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vlanGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "description", ""),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "range", "1-4094"),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "location_id", ""),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vlanGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(vlanGroupResourceName, "created"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccVLANGroupResource_updateAndImport(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vlan-group-update-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccVLANGroupResourceConfig(name, "before", "1-100")},
			{
				Config: testAccVLANGroupResourceConfig(name+"-updated", "after", "101-200"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vlanGroupResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "description", "after"),
					resource.TestCheckResourceAttr(vlanGroupResourceName, "range", "101-200"),
				),
			},
			{ResourceName: vlanGroupResourceName, ImportState: true, ImportStateVerify: true},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccVLANGroupResource_drift(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vlan-group-drift-%d", testutil.AccSeedForTest(t))
	config := testAccVLANGroupResourceConfig(name, "managed by Terraform", "1-4094")
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(vlanGroupResourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "ipam/vlan-groups", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(vlanGroupResourceName, "description", "managed by Terraform")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVLANGroupResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vlan-group-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccVLANGroupResourceConfig(name, "", "1-4094"), Check: testutil.DeleteResourceOutOfBand(vlanGroupResourceName, "ipam/vlan-groups")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVLANGroupResource_parallel(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-vlan-group-parallel-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccVLANGroupResourceParallelConfig(name), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("nautobot_vlan_group.first", "id"),
			resource.TestCheckResourceAttrSet("nautobot_vlan_group.second", "id"),
			resource.TestCheckResourceAttrSet("nautobot_vlan_group.third", "id"),
		)},
		{Config: testutil.AccProviderConfig()},
	}})
}
