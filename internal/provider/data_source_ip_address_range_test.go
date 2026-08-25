package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccIPAddressRangeDataSourceConfig(seed int64, cidr, start, end string) string {
	return testAccIPAddressRangeConfig(seed, cidr, start, end, fmt.Sprintf("tfacc-ds-range-%d", seed), "range data source test", false) + `
data "nautobot_ip_address_range" "test" {
  id = nautobot_ip_address_range.test.id
}
`
}

func TestAccIPAddressRangeDataSource_byID(t *testing.T) {
	t.Parallel()
	seed := testAccSeedForTest(t)
	cidr := testAccPrefixCIDR(seed, 25)
	start, end := testAccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) }, ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeDataSourceConfig(seed, cidr, start, end), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPair("data.nautobot_ip_address_range.test", "id", "nautobot_ip_address_range.test", "id"), resource.TestCheckResourceAttr("data.nautobot_ip_address_range.test", "start_address", start), resource.TestCheckResourceAttr("data.nautobot_ip_address_range.test", "end_address", end), resource.TestCheckResourceAttrSet("data.nautobot_ip_address_range.test", "namespace_id"),
		)}, {Config: testAccProviderConfig()},
	}})
}
