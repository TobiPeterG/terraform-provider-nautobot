package provider

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAvailableIPAddressConfigMinimal(name string, vid int, cidr string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
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
`, name, vid, cidr, testStatus)
}

func testAccAvailableIPAddressConfigFull(name string, vid int, cidr, status, ipType, dnsName, description, roleID string, tagIDs []string, withTenant, withNatInside bool) string {
	roleAttr := ""
	if roleID != "" {
		roleAttr = fmt.Sprintf("  role_id       = %q\n", roleID)
	}
	tenantAttr := ""
	if withTenant {
		tenantAttr = "  tenant_id     = nautobot_tenant.t.id\n"
	}
	natAttr := ""
	if withNatInside {
		natAttr = "  nat_inside_id = nautobot_available_ip_address.inside.id\n"
	}
	tagsAttr := ""
	if tagIDs != nil {
		quoted := make([]string, 0, len(tagIDs))
		for _, id := range tagIDs {
			quoted = append(quoted, fmt.Sprintf("%q", id))
		}
		tagsAttr = fmt.Sprintf("  tags_ids      = [%s]\n", strings.Join(quoted, ", "))
	}

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "t" {
  name = "%[1]s-tenant"
}

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

resource "nautobot_available_ip_address" "inside" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
}

resource "nautobot_available_ip_address" "test" {
  prefix_id   = nautobot_prefix.p.id
  status      = "%[5]s"
  type        = "%[6]s"
  dns_name    = "%[7]s"
  description = "%[8]s"
%[9]s%[10]s%[11]s%[12]s}
`, name, vid, cidr, testStatus, status, ipType, dnsName, description, roleAttr, tenantAttr, natAttr, tagsAttr)
}

func testAccAvailableIPAddressConfigParallel(name string, vid int, cidr string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
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
`, name, vid, cidr, testStatus)
}

func TestAccAvailableIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-min-%d", seed)
	vid := testAccVLANVid(seed, 9)
	cidr := testAccPrefixCIDR(seed, 3)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "address"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttrSet(resourceName, "mask_length"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_version"),
					resource.TestCheckResourceAttr(resourceName, "type", "host"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
					resource.TestCheckResourceAttrSet(resourceName, "display"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrSet(resourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(resourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_update(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-upd-%d", seed)
	vid := testAccVLANVid(seed, 10)
	cidr := testAccPrefixCIDR(seed, 4)

	roleID := testAccCreateRole(t, fmt.Sprintf("%s-role", name), []string{"ipam.ipaddress"})
	tag1ID := testAccCreateTag(t, fmt.Sprintf("%s-tag-a", name), []string{"ipam.ipaddress"})
	tag2ID := testAccCreateTag(t, fmt.Sprintf("%s-tag-b", name), []string{"ipam.ipaddress"})

	resourceName := "nautobot_available_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-ip-%d.example.com", time.Now().Unix())
	dnsName2 := fmt.Sprintf("tfacc-ip-upd-%d.example.com", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigFull(name, vid, cidr, testStatus, "host", "", "", "", nil, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "host"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
				),
			},
			{
				// Pass tags in reverse order to verify unordered Set handling.
				Config: testAccAvailableIPAddressConfigFull(name, vid, cidr, testStatus, "dhcp", dnsName1, "created by terraform acceptance test", roleID, []string{tag2ID, tag1ID}, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "dhcp"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName1),
					resource.TestCheckResourceAttr(resourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
					resource.TestCheckResourceAttr(resourceName, "role_id", roleID),
					resource.TestCheckResourceAttrPair(resourceName, "tenant_id", "nautobot_tenant.t", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "nat_inside_id", "nautobot_available_ip_address.inside", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags_ids.*", tag1ID),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags_ids.*", tag2ID),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigFull(name, vid, cidr, "Reserved", "slaac", dnsName2, "updated description", "", []string{}, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "slaac"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName2),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigFull(name, vid, cidr, "Reserved", "host", "", "", "", nil, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "type", "host"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-imp-%d", seed)
	vid := testAccVLANVid(seed, 11)
	cidr := testAccPrefixCIDR(seed, 5)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigFull(name, vid, cidr, testStatus, "dhcp", fmt.Sprintf("tfacc-imp-%d.example.com", seed), "import test description", "", nil, true, true),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-del-%d", seed)
	vid := testAccVLANVid(seed, 12)
	cidr := testAccPrefixCIDR(seed, 6)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_parallelAllocations(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-par-%d", seed)
	vid := testAccVLANVid(seed, 13)
	cidr := testAccPrefixCIDR(seed, 7)

	resourceName1 := "nautobot_available_ip_address.ip1"
	resourceName2 := "nautobot_available_ip_address.ip2"
	resourceName3 := "nautobot_available_ip_address.ip3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigParallel(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "address"),
					resource.TestCheckResourceAttr(resourceName1, "status", testStatus),

					resource.TestCheckResourceAttrPair(resourceName2, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "address"),
					resource.TestCheckResourceAttr(resourceName2, "status", testStatus),

					resource.TestCheckResourceAttrPair(resourceName3, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "address"),
					resource.TestCheckResourceAttr(resourceName3, "status", testStatus),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
