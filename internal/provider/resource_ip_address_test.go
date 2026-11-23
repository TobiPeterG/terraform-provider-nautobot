package provider

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccIPFromPrefixCIDR(t *testing.T, prefixCIDR string, hostOffset int) string {
	t.Helper()
	p, err := netip.ParsePrefix(prefixCIDR)
	if err != nil {
		t.Fatalf("parse prefix %q: %v", prefixCIDR, err)
	}
	addr := p.Addr()
	for i := 0; i < hostOffset; i++ {
		addr = addr.Next()
	}
	return fmt.Sprintf("%s/%d", addr.String(), p.Bits())
}

func testAccIPAddressConfigMinimal(name string, vid int, cidr, ipCIDR string) string {
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

resource "nautobot_ip_address" "test" {
  address    = "%[5]s"
  status     = "%[4]s"
  depends_on = [nautobot_prefix.p]
}
`, name, vid, cidr, testStatus, ipCIDR)
}

func testAccIPAddressConfigNamespaceWithoutParent(name string, vid int, cidr, ipCIDR, customNamespaceID, description string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "global_p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_ip_address" "test" {
  address      = "%[5]s"
  status       = "%[4]s"
  namespace_id = "%[6]s"
  description  = "%[7]s"
  depends_on   = [nautobot_prefix.global_p]
}
`, name, vid, cidr, testStatus, ipCIDR, customNamespaceID, description)
}

func testAccIPAddressConfigFull(name string, vid int, cidr, insideIP, ipCIDR, status, ipType, dnsName, description, roleID string, tagIDs []string, withParent, withNamespace, withTenant, withNatInside bool) string {
	parentAttr := ""
	if withParent {
		parentAttr = "  parent_id     = nautobot_prefix.p.id\n"
	}
	nsAttr := ""
	if withNamespace {
		nsAttr = "  namespace_id  = nautobot_prefix.p.namespace_id\n"
	}
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
		natAttr = "  nat_inside_id = nautobot_ip_address.inside.id\n"
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

resource "nautobot_ip_address" "inside" {
  address   = "%[5]s"
  status    = "%[4]s"
  parent_id = nautobot_prefix.p.id
}

resource "nautobot_ip_address" "test" {
  address     = "%[6]s"
  status      = "%[7]s"
  type        = "%[8]s"
  dns_name    = "%[9]s"
  description = "%[10]s"
%[11]s%[12]s%[13]s%[14]s%[15]s%[16]s  depends_on  = [nautobot_prefix.p]
}
`, name, vid, cidr, testStatus, insideIP, ipCIDR, status, ipType, dnsName, description, parentAttr, nsAttr, roleAttr, tenantAttr, natAttr, tagsAttr)
}

func TestAccIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-min-%d", seed)
	vid := testAccVLANVid(seed, 31)
	cidr := testAccPrefixCIDR(seed, 31)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigMinimal(name, vid, cidr, ipCIDR),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttrSet(resourceName, "host"),
					resource.TestCheckResourceAttr(resourceName, "mask_length", "28"),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "4"),
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

func TestAccIPAddressResource_namespaceWithoutParent(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-ns-%d", seed)
	vid := testAccVLANVid(seed, 35)
	cidr := testAccPrefixCIDR(seed, 35)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	customNSID, customPrefixID := testAccCreateNamespaceAndPrefix(t, fmt.Sprintf("%s-ns", name), cidr)
	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigNamespaceWithoutParent(name, vid, cidr, ipCIDR, customNSID, "in custom namespace"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR),
					resource.TestCheckResourceAttr(resourceName, "namespace_id", customNSID),
					resource.TestCheckResourceAttr(resourceName, "parent_id", customPrefixID),
					resource.TestCheckResourceAttr(resourceName, "description", "in custom namespace"),
				),
			},
			{
				// Updating an unrelated field while parent_id is omitted must preserve parent_id and namespace_id without drift.
				Config: testAccIPAddressConfigNamespaceWithoutParent(name, vid, cidr, ipCIDR, customNSID, "updated in custom namespace"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "namespace_id", customNSID),
					resource.TestCheckResourceAttr(resourceName, "parent_id", customPrefixID),
					resource.TestCheckResourceAttr(resourceName, "description", "updated in custom namespace"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_update(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-upd-%d", seed)
	vid := testAccVLANVid(seed, 32)
	cidr := testAccPrefixCIDR(seed, 32)
	insideIP := testAccIPFromPrefixCIDR(t, cidr, 1)
	ipCIDR1 := testAccIPFromPrefixCIDR(t, cidr, 2)
	ipCIDR2 := testAccIPFromPrefixCIDR(t, cidr, 3)

	roleID := testAccCreateRole(t, fmt.Sprintf("%s-role", name), []string{"ipam.ipaddress"})
	tag1ID := testAccCreateTag(t, fmt.Sprintf("%s-tag-a", name), []string{"ipam.ipaddress"})
	tag2ID := testAccCreateTag(t, fmt.Sprintf("%s-tag-b", name), []string{"ipam.ipaddress"})

	resourceName := "nautobot_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-static-%d.example.com", time.Now().Unix())
	dnsName2 := fmt.Sprintf("tfacc-static-upd-%d.example.com", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR1, testStatus, "host", "", "", "", nil, true, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR1),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
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
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR2, "Reserved", "dhcp", dnsName1, "created by terraform acceptance test", roleID, []string{tag2ID, tag1ID}, true, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR2),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "type", "dhcp"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName1),
					resource.TestCheckResourceAttr(resourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
					resource.TestCheckResourceAttr(resourceName, "role_id", roleID),
					resource.TestCheckResourceAttrPair(resourceName, "tenant_id", "nautobot_tenant.t", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "nat_inside_id", "nautobot_ip_address.inside", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags_ids.*", tag1ID),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags_ids.*", tag2ID),
				),
			},
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR2, "Reserved", "slaac", dnsName2, "", "", []string{}, false, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR2),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "type", "slaac"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-imp-%d", seed)
	vid := testAccVLANVid(seed, 33)
	cidr := testAccPrefixCIDR(seed, 33)
	insideIP := testAccIPFromPrefixCIDR(t, cidr, 1)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR, testStatus, "dhcp", fmt.Sprintf("tfacc-ip-imp-%d.example.com", seed), "import test description", "", nil, true, false, true, true),
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

func TestAccIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-del-%d", seed)
	vid := testAccVLANVid(seed, 34)
	cidr := testAccPrefixCIDR(seed, 34)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigMinimal(name, vid, cidr, ipCIDR),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
