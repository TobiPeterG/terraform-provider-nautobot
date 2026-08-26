package vlan_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	vlanDataSourceName = "data.nautobot_vlan.test"
)

func TestAccVLANDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vlan-missing-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_vlan" "test" {
  name = %q
}
`, name),
			ExpectError: regexp.MustCompile(`VLAN lookup failed`),
		}},
	})
}

func testAccVLANDataSourceConfigMinimal(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name   = "%[1]s"
  vid    = %[2]d
  status = "%[3]s"
}

data "nautobot_vlan" "test" {
  depends_on = [nautobot_vlan.test]
  id         = nautobot_vlan.test.id
}
`, name, vid, status)
}

func TestAccVLANDataSource_multipleUngroupedMatches(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-ds-vlan-duplicate-%d", seed)
	vid := testutil.AccVLANVID(seed, 9)
	config := testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "one" {
  name = %[1]q
  vid = %[2]d
  status = %[3]q
}
resource "nautobot_vlan" "two" {
  name = %[1]q
  vid = %[2]d + 1
  status = %[3]q
}
data "nautobot_vlan" "test" {
  depends_on = [nautobot_vlan.one, nautobot_vlan.two]
  name = %[1]q
}
`, name, vid, testutil.Status)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`(?s)VLAN lookup failed.*found 2`),
		}},
	})
}

func testAccVLANDataSourceConfigFull(name string, vid int) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan_group" "test" {
  name = "%[1]s-group"
}

resource "nautobot_tenant" "test" {
  name = "%[1]s-tenant"
}

resource "nautobot_vlan" "test" {
  name        = "%[1]s"
  vid         = %[2]d
  status      = "%[3]s"
  description = "created by terraform acceptance test"
  vlan_group_id = nautobot_vlan_group.test.id
  tenant_id     = nautobot_tenant.test.id
}

data "nautobot_vlan" "test" {
  name          = nautobot_vlan.test.name
  vlan_group_id = nautobot_vlan_group.test.id
}
`,
		name,
		vid,
		status,
	)
}

func TestAccVLANDataSource_minimal(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-ds-vlan-minimal-%d", seed)
	vid := testutil.AccVLANVID(seed, 6)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANDataSourceConfigMinimal(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanDataSourceName, "name", name),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanDataSourceName, "status", testutil.Status),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "id"),
					resource.TestCheckResourceAttr(vlanDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vlan_group_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "notes_url"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccVLANDataSource_full(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-ds-vlan-full-%d", seed)
	vid := testutil.AccVLANVID(seed, 7)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANDataSourceConfigFull(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanDataSourceName, "name", name),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanDataSourceName, "status", testutil.Status),

					resource.TestCheckResourceAttr(vlanDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(vlanDataSourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttrPair(vlanDataSourceName, "vlan_group_id", vlanGroupResourceName, "id"),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
