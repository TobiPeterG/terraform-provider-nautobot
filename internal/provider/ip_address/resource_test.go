package ip_address_test

import (
	"fmt"
	"net/netip"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/nautobot/terraform-provider-nautobot/internal/provider/ip_address"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestIPAddressDerivedAttributesRemainUnknownDuringUpdate(t *testing.T) {
	t.Parallel()
	testutil.AssertStringAttributesHaveNoPlanModifiers(
		t,
		NewIPAddressResource(),
		"host",
	)
}

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

resource "nautobot_ip_address" "test" {
  address    = "%[5]s"
  status     = "%[4]s"
  depends_on = [nautobot_prefix.p]
}
`, name, vid, cidr, testutil.Status, ipCIDR)
}

func testAccIPAddressConfigNamespaceWithoutParent(name string, vid int, cidr, ipCIDR, customNamespaceID, description string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
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
`, name, vid, cidr, testutil.Status, ipCIDR, customNamespaceID, description)
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

	return testutil.AccProviderConfig() + fmt.Sprintf(`
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
`, name, vid, cidr, testutil.Status, insideIP, ipCIDR, status, ipType, dnsName, description, parentAttr, nsAttr, roleAttr, tenantAttr, natAttr, tagsAttr)
}

func TestAccIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-min-%d", seed)
	vid := testutil.AccVLANVID(seed, 31)
	cidr := testutil.AccPrefixCIDR(seed, 31)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
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
					resource.TestCheckResourceAttr(resourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_namespaceWithoutParent(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-ns-%d", seed)
	vid := testutil.AccVLANVID(seed, 35)
	cidr := testutil.AccPrefixCIDR(seed, 35)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	customNSID, customPrefixID := testutil.CreateTestNamespaceAndPrefix(t, fmt.Sprintf("%s-ns", name), cidr)
	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
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
				Config: testAccIPAddressConfigNamespaceWithoutParent(name, vid, cidr, ipCIDR, customNSID, "updated in custom namespace"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "namespace_id", customNSID),
					resource.TestCheckResourceAttr(resourceName, "parent_id", customPrefixID),
					resource.TestCheckResourceAttr(resourceName, "description", "updated in custom namespace"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_update(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-upd-%d", seed)
	vid := testutil.AccVLANVID(seed, 32)
	cidr := testutil.AccPrefixCIDR(seed, 32)
	insideIP := testAccIPFromPrefixCIDR(t, cidr, 1)
	ipCIDR1 := testAccIPFromPrefixCIDR(t, cidr, 2)
	ipCIDR2 := testAccIPFromPrefixCIDR(t, cidr, 3)

	roleID := testutil.CreateTestRole(t, fmt.Sprintf("%s-role", name), []string{"ipam.ipaddress"})
	tag1ID := testutil.CreateTestTag(t, fmt.Sprintf("%s-tag-a", name), []string{"ipam.ipaddress"})
	tag2ID := testutil.CreateTestTag(t, fmt.Sprintf("%s-tag-b", name), []string{"ipam.ipaddress"})

	resourceName := "nautobot_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-static-%d.example.com", seed)
	dnsName2 := fmt.Sprintf("tfacc-static-upd-%d.example.com", seed)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR1, testutil.Status, "host", "", "", "", nil, true, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR1),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "type", "host"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(resourceName, "role_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_inside_id", ""),
					resource.TestCheckResourceAttr(resourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR1, "Reserved", "dhcp", dnsName1, "created by terraform acceptance test", roleID, []string{tag2ID, tag1ID}, true, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR1),
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
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR2, "Reserved", "host", dnsName2, "", "", []string{}, false, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "address", ipCIDR2),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_prefix.p", "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "type", "host"),
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
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-drift-%d", seed)
	vid := testutil.AccVLANVID(seed, 36)
	cidr := testutil.AccPrefixCIDR(seed, 36)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)
	config := testAccIPAddressConfigMinimal(name, vid, cidr, ipCIDR)
	var id string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: testutil.CaptureResourceID("nautobot_ip_address.test", &id)},
			{
				PreConfig: func() {
					testutil.MutateResourceOutOfBand(t, id, "ipam/ip-addresses", map[string]any{"description": "out-of-band edit"})
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{Config: config, Check: resource.TestCheckResourceAttr("nautobot_ip_address.test", "description", "")},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccIPAddressResource_import(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-imp-%d", seed)
	vid := testutil.AccVLANVID(seed, 33)
	cidr := testutil.AccPrefixCIDR(seed, 33)
	insideIP := testAccIPFromPrefixCIDR(t, cidr, 1)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resourceName := "nautobot_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigFull(name, vid, cidr, insideIP, ipCIDR, testutil.Status, "dhcp", fmt.Sprintf("tfacc-ip-imp-%d.example.com", seed), "import test description", "", nil, true, false, true, true),
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

func TestAccIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-del-%d", seed)
	vid := testutil.AccVLANVID(seed, 34)
	cidr := testutil.AccPrefixCIDR(seed, 34)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigMinimal(name, vid, cidr, ipCIDR),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccIPAddressResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-ip-del-gone-%d", seed)
	cidr := testutil.AccPrefixCIDR(seed, 37)
	ipCIDR := testAccIPFromPrefixCIDR(t, cidr, 2)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		AdditionalCLIOptions:     &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressConfigMinimal(name, testutil.AccVLANVID(seed, 37), cidr, ipCIDR),
				Check:  testutil.DeleteResourceOutOfBand("nautobot_ip_address.test", "ipam/ip-addresses"),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
