package vlan_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const vlanGroupsDataSourceName = "data.nautobot_vlan_groups.test"

func TestAccVLANGroupsDataSource_list(t *testing.T) {
	t.Parallel()
	base := fmt.Sprintf("tfacc-ds-vlan-groups-%d", testutil.AccSeedForTest(t))
	config := testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan_group" "one" {
  name        = %[1]q
  description = "first group"
  range       = "100-199"
}
resource "nautobot_vlan_group" "two" { name = %[2]q }
resource "nautobot_vlan_group" "three" { name = %[3]q }

data "nautobot_vlan_groups" "test" {
  depends_on = [nautobot_vlan_group.one, nautobot_vlan_group.two, nautobot_vlan_group.three]
}
`, base+"-1", base+"-2", base+"-3")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				testutil.CountAtLeast(vlanGroupsDataSourceName, "vlan_groups", 3),
				testutil.FindListIndexByAttr(vlanGroupsDataSourceName, "vlan_groups", "name", base+"-1"),
				testutil.FindListIndexByAttr(vlanGroupsDataSourceName, "vlan_groups", "name", base+"-2"),
				testutil.FindListIndexByAttr(vlanGroupsDataSourceName, "vlan_groups", "name", base+"-3"),
				testutil.CheckListItemHasAttrs(
					vlanGroupsDataSourceName,
					"vlan_groups",
					"name",
					base+"-1",
					map[string]string{
						"description": "first group",
						"range":       "100-199",
						"location_id": "",
						"tags_ids.#":  "0",
						"vlan_count":  "0",
					},
					[]string{"id", "created", "last_updated", "display", "url", "natural_slug", "notes_url"},
				),
			),
		}, {Config: testutil.AccProviderConfig()}},
	})
}
