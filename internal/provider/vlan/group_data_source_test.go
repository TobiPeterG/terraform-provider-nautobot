package vlan_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const vlanGroupDataSourceName = "data.nautobot_vlan_group.test"

func testAccVLANGroupDataSourceConfig(name string, byID bool) string {
	selector := "name = nautobot_vlan_group.test.name"
	if byID {
		selector = "id = nautobot_vlan_group.test.id"
	}
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan_group" "test" {
  name        = %[1]q
  description = "acceptance test VLAN group"
  range       = "10-99,200"
}

data "nautobot_vlan_group" "test" {
  %[2]s
}
`, name, selector)
}

func TestAccVLANGroupDataSource_byID(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-vlan-group-id-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccVLANGroupDataSourceConfig(name, true),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrPair(vlanGroupDataSourceName, "id", vlanGroupResourceName, "id"),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "name", name),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "description", "acceptance test VLAN group"),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "range", "10-99,200"),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "location_id", ""),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "tags_ids.#", "0"),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "vlan_count", "0"),
				resource.TestCheckResourceAttrSet(vlanGroupDataSourceName, "created"),
				resource.TestCheckResourceAttrSet(vlanGroupDataSourceName, "last_updated"),
			),
		}, {Config: testutil.AccProviderConfig()}},
	})
}

func TestAccVLANGroupDataSource_byName(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-vlan-group-name-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testAccVLANGroupDataSourceConfig(name, false),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrPair(vlanGroupDataSourceName, "id", vlanGroupResourceName, "id"),
				resource.TestCheckResourceAttr(vlanGroupDataSourceName, "name", name),
			),
		}, {Config: testutil.AccProviderConfig()}},
	})
}

func TestAccVLANGroupDataSource_notFound(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-vlan-group-missing-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      testutil.AccProviderConfig() + fmt.Sprintf("\ndata \"nautobot_vlan_group\" \"test\" { name = %q }\n", name),
			ExpectError: regexp.MustCompile(`VLAN group lookup failed`),
		}},
	})
}
