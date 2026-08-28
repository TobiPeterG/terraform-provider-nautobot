package ip_address_range_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestAccIPAddressRangeDataSource_notFound(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 52)
	start, end := testutil.AccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "missing_range" {
  name = "tf-acc-ds-range-missing-%d"
}

data "nautobot_ip_address_range" "test" {
  start_address = %q
  end_address   = %q
  namespace_id  = nautobot_namespace.missing_range.id
}
`, seed, start, end),
			ExpectError: regexp.MustCompile(`IP address range lookup failed`),
		}},
	})
}

func testAccIPAddressRangeDataSourceConfig(seed int64, cidr, start, end string) string {
	return testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-ds-range-%d", seed), "range data source test", false, false) + `
data "nautobot_ip_address_range" "test" {
  id = nautobot_ip_address_range.test.id
}

`
}

func testAccIPAddressRangeDataSourceNaturalKeyConfig(seed int64, cidr, start, end string) string {
	return testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-ds-range-natural-%d", seed), "range natural-key data source test", false, false) + `
data "nautobot_ip_address_range" "test" {
  start_address = nautobot_ip_address_range.test.start_address
  end_address   = nautobot_ip_address_range.test.end_address
  namespace_id  = nautobot_prefix.p.namespace_id
}
`
}

func TestAccIPAddressRangeDataSource_byID(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 35)
	start, end := testutil.AccIPRangeBounds(cidr)
	dataSourceName := "data.nautobot_ip_address_range.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressRangeDataSourceConfig(seed, cidr, start, end),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "nautobot_ip_address_range.test", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "start_address", start),
					resource.TestCheckResourceAttr(dataSourceName, "end_address", end),
					resource.TestCheckResourceAttrSet(dataSourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "description", "range data source test"),
					resource.TestCheckResourceAttr(dataSourceName, "count_as_utilized", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "is_exclusive", "false"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccIPAddressRangeDataSource_byBoundariesAndNamespace(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 48)
	start, end := testutil.AccIPRangeBounds(cidr)
	dataSourceName := "data.nautobot_ip_address_range.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressRangeDataSourceNaturalKeyConfig(seed, cidr, start, end),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "nautobot_ip_address_range.test", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "start_address", start),
					resource.TestCheckResourceAttr(dataSourceName, "end_address", end),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttr(dataSourceName, "description", "range natural-key data source test"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
