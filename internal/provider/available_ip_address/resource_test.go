package available_ip_address_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/nautobot/terraform-provider-nautobot/internal/provider/available_ip_address"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestAvailableIPAddressDerivedAttributesRemainUnknownDuringReplacement(t *testing.T) {
	t.Parallel()
	testutil.AssertStringAttributesHaveNoPlanModifiers(t, NewAvailableIPAddressResource(), "address")
}

func testAccAvailableIPAddressConfigMinimal(name string, vid int, cidr string) string {
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

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
}
`, name, vid, cidr, testutil.Status)
}

func testAccAvailableIPAddressConfigWithDNSName(name string, vid int, cidr string, dnsName string) string {
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

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[5]s"
}
`, name, vid, cidr, testutil.Status, dnsName)
}

func testAccAvailableIPAddressConfigWithStatusAndDNSName(name string, vid int, cidr string, status, dnsName string) string {
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

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
  dns_name  = "%[6]s"
}
`, name, vid, cidr, testutil.Status, status, dnsName)
}

func testAccAvailableIPAddressConfigParallel(name string, vid int, cidr string) string {
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

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-1"
}

resource "nautobot_available_ip_address" "ip2" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-2"
}

resource "nautobot_available_ip_address" "ip3" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-3"
}
`, name, vid, cidr, testutil.Status)
}

func TestAccAvailableIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-min-%d", seed)
	vid := testutil.AccVLANVID(seed, 9)
	cidr := testutil.AccPrefixCIDR(seed, 3)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "address"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_version"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testutil.Status),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_update(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-upd-%d", seed)
	vid := testutil.AccVLANVID(seed, 10)
	cidr := testutil.AccPrefixCIDR(seed, 4)

	resourceName := "nautobot_available_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-ip-%d.example.com", seed)
	dnsName2 := fmt.Sprintf("tfacc-ip-upd-%d.example.com", seed)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testutil.Status),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithDNSName(name, vid, cidr, dnsName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName1),
					resource.TestCheckResourceAttr(resourceName, "status", testutil.Status),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName(name, vid, cidr, "Reserved", dnsName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName2),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName(name, vid, cidr, "Reserved", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-drift-%d", seed)
	dnsName := fmt.Sprintf("tfacc-ip-drift-%d.example.com", seed)
	config := testAccAvailableIPAddressConfigWithDNSName(name, testutil.AccVLANVID(seed, 65), testutil.AccPrefixCIDR(seed, 65), dnsName)
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID("nautobot_available_ip_address.test", &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "ipam/ip-addresses", map[string]any{"dns_name": "outside-terraform.example.com"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr("nautobot_available_ip_address.test", "dns_name", dnsName)},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccAvailableIPAddressResource_import(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-imp-%d", seed)
	vid := testutil.AccVLANVID(seed, 11)
	cidr := testutil.AccPrefixCIDR(seed, 5)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-del-%d", seed)
	vid := testutil.AccVLANVID(seed, 12)
	cidr := testutil.AccPrefixCIDR(seed, 6)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-del-gone-%d", seed)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccAvailableIPAddressConfigMinimal(name, testutil.AccVLANVID(seed, 64), testutil.AccPrefixCIDR(seed, 64)), Check: testutil.DeleteResourceOutOfBand("nautobot_available_ip_address.test", "ipam/ip-addresses")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccAvailableIPAddressResource_parallelAllocations(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-par-%d", seed)
	vid := testutil.AccVLANVID(seed, 13)
	cidr := testutil.AccPrefixCIDR(seed, 7)

	resourceName1 := "nautobot_available_ip_address.ip1"
	resourceName2 := "nautobot_available_ip_address.ip2"
	resourceName3 := "nautobot_available_ip_address.ip3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigParallel(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "address"),
					resource.TestCheckResourceAttr(resourceName1, "status", testutil.Status),

					resource.TestCheckResourceAttrPair(resourceName2, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "address"),
					resource.TestCheckResourceAttr(resourceName2, "status", testutil.Status),

					resource.TestCheckResourceAttrPair(resourceName3, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "address"),
					resource.TestCheckResourceAttr(resourceName3, "status", testutil.Status),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func testAccAvailableIPAddressFromRangeConfig(cidr, start, end string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
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
resource "nautobot_available_ip_address" "test" {
  ip_address_range_id = nautobot_ip_address_range.r.id
  status              = %q
}
`, cidr, testutil.Status, start, end, testutil.Status, testutil.Status)
}

func TestAccAvailableIPAddressResource_fromIPRange(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 26)
	start, end := testutil.AccAvailableIPRangeBounds(cidr)
	expectedAddress := start + cidr[strings.LastIndex(cidr, "/"):]
	addr := "nautobot_available_ip_address.test"
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccAvailableIPAddressFromRangeConfig(cidr, start, end), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttrPair(addr, "ip_address_range_id", "nautobot_ip_address_range.r", "id"), resource.TestCheckResourceAttrPair(addr, "prefix_id", "nautobot_prefix.p", "id"), resource.TestCheckResourceAttr(addr, "address", expectedAddress), resource.TestCheckResourceAttr(addr, "status", testutil.Status))},
		{Config: testutil.AccProviderConfig()},
	}})
}

func testAccAvailableIPAddressRangeReplacementConfig(firstCIDR, firstStart, firstEnd, secondCIDR, secondStart, secondEnd, selectedRange string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "first" {
  prefix = %q
  status = %q
}

resource "nautobot_prefix" "second" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "first" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.first.id
  status        = %q
}

resource "nautobot_ip_address_range" "second" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.second.id
  status        = %q
}

resource "nautobot_available_ip_address" "test" {
  ip_address_range_id = nautobot_ip_address_range.%s.id
  status              = %q
}
`, firstCIDR, testutil.Status, secondCIDR, testutil.Status,
		firstStart, firstEnd, testutil.Status,
		secondStart, secondEnd, testutil.Status,
		selectedRange, testutil.Status)
}

func TestAccAvailableIPAddressResource_replaceAllocationRange(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	firstCIDR := testutil.AccPrefixCIDR(seed, 41)
	secondCIDR := testutil.AccPrefixCIDR(seed, 42)
	firstStart, firstEnd := testutil.AccAvailableIPRangeBounds(firstCIDR)
	secondStart, secondEnd := testutil.AccAvailableIPRangeBounds(secondCIDR)
	resourceName := "nautobot_available_ip_address.test"

	config := func(selectedRange string) string {
		return testAccAvailableIPAddressRangeReplacementConfig(
			firstCIDR, firstStart, firstEnd,
			secondCIDR, secondStart, secondEnd,
			selectedRange,
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "ip_address_range_id", "nautobot_ip_address_range.first", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.first", "id"),
					resource.TestCheckResourceAttr(resourceName, "address", firstStart+firstCIDR[strings.LastIndex(firstCIDR, "/"):]),
				),
			},
			{
				Config: config("second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "ip_address_range_id", "nautobot_ip_address_range.second", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.second", "id"),
					resource.TestCheckResourceAttr(resourceName, "address", secondStart+secondCIDR[strings.LastIndex(secondCIDR, "/"):]),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func testAccAvailableIPAddressFullRangeConfig(cidr, address string, allocateSecond bool) string {
	second := ""
	if allocateSecond {
		second = fmt.Sprintf(`
resource "nautobot_available_ip_address" "second" {
  ip_address_range_id = nautobot_ip_address_range.r.id
  status              = %q
  depends_on          = [nautobot_available_ip_address.first]
}
`, testutil.Status)
	}
	return testutil.AccProviderConfig() + fmt.Sprintf(`
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
resource "nautobot_available_ip_address" "first" {
  ip_address_range_id = nautobot_ip_address_range.r.id
  status              = %q
}
%s
`, cidr, testutil.Status, address, address, testutil.Status, testutil.Status, second)
}

func TestAccAvailableIPAddressResource_fullIPRangeDoesNotFallBackToPrefix(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 28)
	start, _ := testutil.AccAvailableIPRangeBounds(cidr)
	expectedAddress := start + cidr[strings.LastIndex(cidr, "/"):]
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccAvailableIPAddressFullRangeConfig(cidr, start, false), Check: resource.TestCheckResourceAttr("nautobot_available_ip_address.first", "address", expectedAddress)},
		{Config: testAccAvailableIPAddressFullRangeConfig(cidr, start, true), ExpectError: regexp.MustCompile(`(?i)(failed to allocate IP address|no IP address id returned|no available IP)`)},
		{Config: testutil.AccProviderConfig()},
	}})
}
