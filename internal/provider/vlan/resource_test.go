package vlan_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	vlanResourceName = "nautobot_vlan.test"
)

func testAccVLANConfigMinimal(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name   = "%s"
  vid    = %d
  status = "%s"
}
`, name, vid, status)
}

func testAccVLANConfigFull(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name = "%s-tenant"
}

resource "nautobot_vlan_group" "test" {
  name = "%s-group"
}

resource "nautobot_vlan" "test" {
  name        = "%s"
  vid         = %d
  status      = "%s"

  description = "created by terraform acceptance test"
  tenant_id   = nautobot_tenant.test.id
  vlan_group_id = nautobot_vlan_group.test.id
}
`, name, name, name, vid, status)
}

func testAccVLANConfigUpdated(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name = "%s-tenant"
}

resource "nautobot_vlan_group" "test" {
  name = "%s-group"
}

resource "nautobot_vlan" "test" {
  name        = "%s-updated"
  vid         = %d
  status      = "%s"

  description = "updated by terraform acceptance test"
}
`, name, name, name, vid, status)
}

func testAccVLANConfigUpdatedWithoutDependencies(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name        = "%s-updated"
  vid         = %d
  status      = "%s"

  description = "updated by terraform acceptance test"
}
`, name, vid, status)
}

func testAccVLANConfigParallel(baseName string, baseVid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "vlan1" {
  name   = "%s-1"
  vid    = %d
  status = "%s"
}

resource "nautobot_vlan" "vlan2" {
  name   = "%s-2"
  vid    = %d
  status = "%s"
}

resource "nautobot_vlan" "vlan3" {
  name   = "%s-3"
  vid    = %d
  status = "%s"
}
`, baseName, baseVid, status, baseName, baseVid+1, status, baseName, baseVid+2, status)
}

func TestAccVLANResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-minimal-%d", seed)
	vid := testutil.AccVLANVID(seed, 26)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanResourceName, "status", testutil.Status),

					resource.TestCheckResourceAttr(vlanResourceName, "description", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "vlan_group_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_full(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-full-%d", seed)
	vid := testutil.AccVLANVID(seed, 27)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigFull(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanResourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(vlanResourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttrPair(vlanResourceName, "vlan_group_id", "nautobot_vlan_group.test", "id"),
					resource.TestCheckResourceAttr(vlanResourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_update(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-update-%d", seed)

	vid1 := testutil.AccVLANVID(seed, 28)
	vid2 := testutil.AccVLANVID(seed, 29)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigFull(name, vid1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid1)),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccVLANConfigUpdated(name, vid2),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVLANConfigUpdated(name, vid2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid2)),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(vlanResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "vlan_group_id", ""),
				),
			},
			{Config: testAccVLANConfigUpdatedWithoutDependencies(name, vid2)},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vlan-drift-%d", seed)
	config := testAccVLANConfigFull(name, testutil.AccVLANVID(seed, 72))
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(vlanResourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "ipam/vlans", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(vlanResourceName, "description", "created by terraform acceptance test")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVLANResource_import(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-import-%d", seed)
	vid := testutil.AccVLANVID(seed, 30)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
			},
			{
				ResourceName:      vlanResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_delete(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-delete-%d", seed)
	vid := testutil.AccVLANVID(seed, 31)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vlan-del-gone-%d", seed)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccVLANConfigMinimal(name, testutil.AccVLANVID(seed, 63)), Check: testutil.DeleteResourceOutOfBand(vlanResourceName, "ipam/vlans")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccVLANResource_parallel(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	baseName := fmt.Sprintf("tf-acc-vlan-parallel-%d", seed)

	baseVid := testutil.AccVLANVID(seed, 32)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigParallel(baseName, baseVid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "name", baseName+"-1"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "vid", fmt.Sprintf("%d", baseVid)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "name", baseName+"-2"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "vid", fmt.Sprintf("%d", baseVid+1)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "name", baseName+"-3"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "vid", fmt.Sprintf("%d", baseVid+2)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
