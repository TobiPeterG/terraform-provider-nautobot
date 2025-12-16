package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	prefixResourceName = "nautobot_prefix.test"
)

func testAccPrefixCIDR(seed int64, offset int) string {
	oct3 := int(seed % 200)
	oct4 := offset * 16
	if oct4 > 240 {
		oct4 = 240
	}
	return fmt.Sprintf("10.200.%d.%d/28", oct3, oct4)
}

func testAccPrefixConfigMinimal(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigFull(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "created by terraform acceptance test"
  tenant_id   = "%[5]s"
}
`, name, vid, status, cidr, testTenantID)
}

func testAccPrefixConfigUpdated(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "updated by terraform acceptance test"
  tenant_id   = "%[5]s"
}
`, name, vid, status, cidr, testTenantID)
}

func testAccPrefixConfigParallel(baseName string, baseVid int, seed int64) string {
	status := testStatus
	c1 := testAccPrefixCIDR(seed, 1)
	c2 := testAccPrefixCIDR(seed, 2)
	c3 := testAccPrefixCIDR(seed, 3)

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p1" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[5]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[6]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, baseName, baseVid, status, c1, c2, c3)
}

func TestAccPrefixResource_minimal(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	name := fmt.Sprintf("tf-acc-prefix-minimal-%d", seed)
	vid := testAccVLANVid(seed, 60)
	cidr := testAccPrefixCIDR(seed, 0)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testStatus),

					resource.TestCheckResourceAttrSet(prefixResourceName, "id"),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr(prefixResourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "rir_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "date_allocated", ""),

					resource.TestCheckResourceAttr(prefixResourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(prefixResourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "ip_version"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "display"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "url"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_full(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	name := fmt.Sprintf("tf-acc-prefix-full-%d", seed)
	vid := testAccVLANVid(seed, 70)
	cidr := testAccPrefixCIDR(seed, 10)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testStatus),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", testTenantID),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	name := fmt.Sprintf("tf-acc-prefix-update-%d", seed)
	vid := testAccVLANVid(seed, 80)
	cidr := testAccPrefixCIDR(seed, 20)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccPrefixConfigUpdated(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccPrefixConfigUpdated(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", testTenantID),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_import(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	name := fmt.Sprintf("tf-acc-prefix-import-%d", seed)
	vid := testAccVLANVid(seed, 90)
	cidr := testAccPrefixCIDR(seed, 30)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				ResourceName:      prefixResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_delete(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	name := fmt.Sprintf("tf-acc-prefix-delete-%d", seed)
	vid := testAccVLANVid(seed, 95)
	cidr := testAccPrefixCIDR(seed, 40)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_parallel(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	baseName := fmt.Sprintf("tf-acc-prefix-parallel-%d", seed)
	baseVid := testAccVLANVid(seed, 96)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigParallel(baseName, baseVid, seed),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("nautobot_prefix.p1", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p2", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p3", "id"),

					resource.TestCheckResourceAttrPair("nautobot_prefix.p1", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p2", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p3", "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "status", testStatus),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "status", testStatus),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "status", testStatus),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "description", ""),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
