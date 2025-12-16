package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	prefixDataSourceName = "data.nautobot_prefix.test"
)

func testAccPrefixDataSourceConfigByID() string {
	return testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_prefix" "test" {
  id = "%[1]s"
}
`, testPrefixID)
}

func testAccPrefixDataSourceConfigByVLAN(prefixCIDR string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "test" {
  prefix    = "%[1]s"
  status    = "%[2]s"
  vlan_id   = "%[3]s"
  tenant_id = "%[4]s"
}

data "nautobot_prefix" "test" {
  depends_on = [nautobot_prefix.test]
  vlan_id = nautobot_prefix.test.vlan_id
}
`, prefixCIDR, testStatus, testVLAN, testTenantID)
}

func TestAccPrefixDataSource_byID(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixDataSourceConfigByID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixDataSourceName, "id", testPrefixID),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "last_updated"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "ip_version"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "notes_url"),

					resource.TestCheckResourceAttr(prefixDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "rir_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "namespace_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "vlan_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "date_allocated", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixDataSource_byVLAN(t *testing.T) {
	t.Parallel()

	prefixCIDR := fmt.Sprintf("10.250.%d.0/24", int(time.Now().Unix()%200)+1)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixDataSourceConfigByVLAN(prefixCIDR),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "id"),
					resource.TestCheckResourceAttr(prefixDataSourceName, "vlan_id", testVLAN),
					resource.TestCheckResourceAttr(prefixDataSourceName, "prefix", prefixCIDR),

					resource.TestCheckResourceAttr(prefixDataSourceName, "tenant_id", testTenantID),

					resource.TestCheckResourceAttr(prefixDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "rir_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "namespace_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "date_allocated", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "last_updated"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "ip_version"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
