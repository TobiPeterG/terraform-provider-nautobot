package prefix_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	prefixesDataSourceName = "data.nautobot_prefixes.test"
)

func testAccPrefixesDataSourceConfigBasic(base string, p1 string, p2 string, p3 string) string {

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p1" {
  prefix  = "%[2]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[3]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[4]s"
  status  = "%[5]s"
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, testutil.Status)
}

func testAccPrefixesDataSourceConfigFull(base string, vid int, p1 string, p2 string, p3 string) string {

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[7]d
  status = "%[6]s"
}

# minimal prefix
resource "nautobot_prefix" "p1" {
  prefix = "%[2]s"
  status = "%[6]s"
}

# full prefix A
resource "nautobot_prefix" "p2" {
  prefix      = "%[3]s"
  status      = "%[6]s"
  description = "p2 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = nautobot_vlan.v.id
}

# full prefix B (different values)
resource "nautobot_prefix" "p3" {
  prefix      = "%[4]s"
  status      = "%[6]s"
  description = "p3 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = nautobot_vlan.v.id
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, "", testutil.Status, vid)
}

func TestAccPrefixesDataSource_basic(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-prefixes-basic-%d", testutil.AccSeedForTest(t))
	seed := testutil.AccSeedForTest(t)

	p1 := testutil.AccPrefixCIDR(seed, 24)
	p2 := testutil.AccPrefixCIDR(seed, 25)
	p3 := testutil.AccPrefixCIDR(seed, 36)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigBasic(base, p1, p2, p3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(prefixesDataSourceName, "prefixes", 3),

					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p1),
					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p2),
					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p3),

					testutil.CheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":      p1,
						"description": "",
						"tenant_id":   "",
						"role_id":     "",
						"parent_id":   "",
						"rir_id":      "",
						"vlan_id":     "",
						"tags_ids.#":  "0",
					}),
					testutil.CheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p1, "date_allocated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixesDataSource_full(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	base := fmt.Sprintf("tfacc-ds-prefixes-full-%d", seed)
	vid := testutil.AccVLANVID(seed, 5)

	p1 := testutil.AccPrefixCIDR(seed, 29)
	p2 := testutil.AccPrefixCIDR(seed, 30)
	p3 := testutil.AccPrefixCIDR(seed, 38)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigFull(base, vid, p1, p2, p3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(prefixesDataSourceName, "prefixes", 3),

					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p1),
					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p2),
					testutil.FindListIndexByAttr(prefixesDataSourceName, "prefixes", "prefix", p3),

					testutil.CheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":      p1,
						"description": "",
						"tenant_id":   "",
						"vlan_id":     "",
						"tags_ids.#":  "0",
					}),
					testutil.CheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p1, "date_allocated"),

					testutil.CheckPrefixInListHasAttrs(prefixesDataSourceName, p2, map[string]string{
						"prefix":      p2,
						"description": "p2 created by terraform acceptance test",
						"tenant_id":   "",
						"tags_ids.#":  "0",
					}),
					testutil.CheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p2, "date_allocated"),
					testutil.CheckListItemAttrEqualsResourceAttr(prefixesDataSourceName, "prefixes", "prefix", p2, "vlan_id", "nautobot_vlan.v", "id"),

					testutil.CheckPrefixInListHasAttrs(prefixesDataSourceName, p3, map[string]string{
						"prefix":      p3,
						"description": "p3 created by terraform acceptance test",
						"tenant_id":   "",
						"tags_ids.#":  "0",
					}),
					testutil.CheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p3, "date_allocated"),
					testutil.CheckListItemAttrEqualsResourceAttr(prefixesDataSourceName, "prefixes", "prefix", p3, "vlan_id", "nautobot_vlan.v", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
