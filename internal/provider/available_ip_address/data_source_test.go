package available_ip_address_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	availableIPDataSourceName = "data.nautobot_available_ip_address.test"
)

func testAccAvailableIPAddressDataSourceConfig(seed int64, vid int, cidr string) string {
	name := fmt.Sprintf("tfacc-ds-availip-%d", seed)

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

data "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
}
`, name, vid, cidr, testutil.Status)
}

func testAccAvailableIPAddressDataSourceConfigParallelAfter(seed int64, vid int, cidr string) string {
	name := fmt.Sprintf("tfacc-ds-availip-%d", seed)

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[5]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]d-1"
}

resource "nautobot_available_ip_address" "ip2" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]d-2"
}

resource "nautobot_available_ip_address" "ip3" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]d-3"
}

resource "nautobot_available_ip_address" "ip4" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]d-4"
}

data "nautobot_available_ip_address" "test" {
  prefix_id  = nautobot_prefix.p.id
  depends_on = [
    nautobot_available_ip_address.ip1,
    nautobot_available_ip_address.ip2,
    nautobot_available_ip_address.ip3,
    nautobot_available_ip_address.ip4,
  ]
}
`, seed, vid, cidr, testutil.Status, name)
}

func TestAccAvailableIPAddressDataSource_basic(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	vid := testutil.AccVLANVID(seed, 1)
	cidr := testutil.AccPrefixCIDR(seed, 1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressDataSourceConfig(seed, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(availableIPDataSourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(availableIPDataSourceName, "address"),
					resource.TestCheckResourceAttrSet(availableIPDataSourceName, "ip_version"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressDataSource_parallelAndRefreshWhileAllocating(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	vid := testutil.AccVLANVID(seed, 2)
	cidr := testutil.AccPrefixCIDR(seed, 2)

	ip1 := "nautobot_available_ip_address.ip1"
	ip2 := "nautobot_available_ip_address.ip2"
	ip3 := "nautobot_available_ip_address.ip3"
	ip4 := "nautobot_available_ip_address.ip4"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressDataSourceConfigParallelAfter(seed, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(availableIPDataSourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(availableIPDataSourceName, "address"),
					resource.TestCheckResourceAttrSet(availableIPDataSourceName, "ip_version"),

					func(s *terraform.State) error {
						a1 := s.RootModule().Resources[ip1].Primary.Attributes["address"]
						a2 := s.RootModule().Resources[ip2].Primary.Attributes["address"]
						a3 := s.RootModule().Resources[ip3].Primary.Attributes["address"]
						a4 := s.RootModule().Resources[ip4].Primary.Attributes["address"]

						return testutil.CheckDataSourceAddressNotEqualAllocated(availableIPDataSourceName, a1, a2, a3, a4)(s)
					},

					testutil.CheckDataSourceAddressNotInAllocatedResources(availableIPDataSourceName, ip1, ip2, ip3, ip4),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressDataSource_fromIPRange(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 27)
	start, end := testutil.AccAvailableIPRangeBounds(cidr)
	expectedAddress := start + cidr[strings.LastIndex(cidr, "/"):]
	config := testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}
resource "nautobot_ip_address_range" "r" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.p.id
  status        = %q
  is_exclusive  = false
}
data "nautobot_available_ip_address" "test" {
  ip_address_range_id = nautobot_ip_address_range.r.id
}
`, cidr, testutil.Status, start, end, testutil.Status)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{{Config: config, Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrPair(availableIPDataSourceName, "ip_address_range_id", "nautobot_ip_address_range.r", "id"), resource.TestCheckResourceAttrPair(availableIPDataSourceName, "prefix_id", "nautobot_prefix.p", "id"), resource.TestCheckResourceAttr(availableIPDataSourceName, "address", expectedAddress))}, {Config: testutil.AccProviderConfig()}}})
}

func TestAccAvailableIPAddressDataSource_exhaustedIPRange(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 29)
	start, _ := testutil.AccAvailableIPRangeBounds(cidr)
	config := testAccAvailableIPAddressFullRangeConfig(cidr, start, false) + `
data "nautobot_available_ip_address" "test" {
  ip_address_range_id = nautobot_ip_address_range.r.id
  depends_on          = [nautobot_available_ip_address.first]
}
`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`No available IPs`),
		}},
	})
}
