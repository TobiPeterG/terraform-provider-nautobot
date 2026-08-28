package ip_address_range_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/nautobot/terraform-provider-nautobot/internal/provider/ip_address_range"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestIPAddressRangeResourceSchema(t *testing.T) {
	t.Parallel()
	var resp frameworkresource.SchemaResponse
	(&IPAddressRangeResource{}).Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	for _, name := range []string{"start_address", "end_address", "namespace_id", "parent_id", "name", "description", "count_as_utilized", "is_exclusive", "status", "role_id", "tenant_id", "tags_ids", "size", "ip_version"} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("IP address range resource schema is missing %q", name)
		}
	}
	if _, ok := resp.Schema.Attributes["last_updated"]; ok {
		t.Error("IP address range resource must not expose last_updated; managed resources omit it")
	}
	testutil.AssertStringAttributesHaveNoPlanModifiers(t, NewIPAddressRangeResource(), "namespace_id", "parent_id", "start_host", "end_host")
	namespaceAttribute := resp.Schema.Attributes["namespace_id"].(rschema.StringAttribute)
	if len(namespaceAttribute.Validators) == 0 {
		t.Fatal("namespace_id must validate that namespace_id or parent_id is configured")
	}
}

func testAccIPAddressRangeConfig(cidr, start, end, name, description string, utilized, exclusive bool) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "test" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = %q
  parent_id         = nautobot_prefix.p.id
  description       = %q
  count_as_utilized = %t
  is_exclusive      = %t
  tenant_id         = nautobot_tenant.test.id
}

resource "nautobot_tenant" "test" {
  name = %q
}
`, cidr, testutil.Status, start, end, testutil.Status, name, description, utilized, exclusive, name+"-tenant")
}

func testAccIPAddressRangeMinimalConfig(cidr, start, end string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "test" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.p.id
  status        = %q
}
`, cidr, testutil.Status, start, end, testutil.Status)
}

func testAccIPAddressRangeMinimalConfigWithTenant(cidr, start, end, tenantName string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_tenant" "test" {
  name = %q
}

resource "nautobot_ip_address_range" "test" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.p.id
  status        = %q
}
`, cidr, testutil.Status, tenantName, start, end, testutil.Status)
}

func testAccIPAddressRangeParallelConfig(cidrs, starts, ends []string, seed int64) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "first" {
  prefix = %q
  status = %q
}
resource "nautobot_prefix" "second" {
  prefix = %q
  status = %q
}
resource "nautobot_prefix" "third" {
  prefix = %q
  status = %q
}
resource "nautobot_ip_address_range" "first" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.first.id
  status        = %q
  name          = "tfacc-range-parallel-%d-1"
}
resource "nautobot_ip_address_range" "second" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.second.id
  status        = %q
  name          = "tfacc-range-parallel-%d-2"
}
resource "nautobot_ip_address_range" "third" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.third.id
  status        = %q
  name          = "tfacc-range-parallel-%d-3"
}
`, cidrs[0], testutil.Status, cidrs[1], testutil.Status, cidrs[2], testutil.Status,
		starts[0], ends[0], testutil.Status, seed,
		starts[1], ends[1], testutil.Status, seed,
		starts[2], ends[2], testutil.Status, seed)
}

func testAccIPAddressRangeMissingNamespaceAndParentConfig(start, end string, seed int64) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_ip_address_range" "test_missing_namespace_and_parent" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = "tfacc-range-missing-namespace-and-parent-%d"
  description       = "created by terraform"
  count_as_utilized = false
}
`, start, end, testutil.Status, seed)
}

func testAccIPAddressRangeProvideParentAndNamespaceConfig(cidr, start, end string, seed int64) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "test_parent_and_namespace" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = "tfacc-range-parent-and-namespace-%d"
  description       = "created by terraform"
  count_as_utilized = false
  parent_id         = nautobot_prefix.p.id
  namespace_id      = nautobot_prefix.p.namespace_id
}
`, cidr, testutil.Status, start, end, testutil.Status, seed)
}

func testAccIPAddressRangeProvideOnlyNamespaceConfig(cidr, start, end string, seed int64) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "test_only_namespace" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = "tfacc-range-only-namespace-%d"
  description       = "created by terraform"
  count_as_utilized = false
  namespace_id      = nautobot_prefix.p.namespace_id
}
`, cidr, testutil.Status, start, end, testutil.Status, seed)
}

func TestAccIPAddressRangeResource_lifecycle(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 34)
	start, end := testutil.AccIPRangeBounds(cidr)
	addr := "nautobot_ip_address_range.test"
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-range-%d", seed), "created by terraform", false, false), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet(addr, "id"), resource.TestCheckResourceAttrPair(addr, "parent_id", "nautobot_prefix.p", "id"), resource.TestCheckResourceAttrPair(addr, "tenant_id", "nautobot_tenant.test", "id"), resource.TestCheckResourceAttr(addr, "start_address", start), resource.TestCheckResourceAttr(addr, "end_address", end), resource.TestCheckResourceAttr(addr, "size", "3"), resource.TestCheckResourceAttr(addr, "ip_version", "4"), resource.TestCheckResourceAttr(addr, "count_as_utilized", "false"), resource.TestCheckResourceAttr(addr, "is_exclusive", "false"), resource.TestCheckResourceAttr(addr, "tags_ids.#", "0"), resource.TestCheckResourceAttrSet(addr, "namespace_id"), resource.TestCheckResourceAttrSet(addr, "created"),
		)},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_minimalAndUpdate(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 73)
	start, end := testutil.AccIPRangeBounds(cidr)
	minimal := testAccIPAddressRangeMinimalConfig(cidr, start, end)
	full := testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-range-update-%d", seed), "managed by Terraform", true, true)
	minimalWithTenant := testAccIPAddressRangeMinimalConfigWithTenant(cidr, start, end, fmt.Sprintf("tfacc-range-update-%d-tenant", seed))
	const resourceName = "nautobot_ip_address_range.test"

	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: minimal, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "name", ""),
			resource.TestCheckResourceAttr(resourceName, "description", ""),
			resource.TestCheckResourceAttr(resourceName, "count_as_utilized", "false"),
			resource.TestCheckResourceAttr(resourceName, "is_exclusive", "false"),
		)},
		{Config: full, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tfacc-range-update-%d", seed)),
			resource.TestCheckResourceAttr(resourceName, "description", "managed by Terraform"),
			resource.TestCheckResourceAttr(resourceName, "count_as_utilized", "true"),
			resource.TestCheckResourceAttr(resourceName, "is_exclusive", "true"),
		)},
		{Config: minimalWithTenant, Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "name", ""),
			resource.TestCheckResourceAttr(resourceName, "description", ""),
			resource.TestCheckResourceAttr(resourceName, "count_as_utilized", "false"),
			resource.TestCheckResourceAttr(resourceName, "is_exclusive", "false"),
			resource.TestCheckResourceAttr(resourceName, "tenant_id", ""),
		)},
		{Config: minimal},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 66)
	start, end := testutil.AccIPRangeBounds(cidr)
	config := testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-range-drift-%d", seed), "managed by Terraform", false, false)
	const resourceName = "nautobot_ip_address_range.test"
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(resourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "ipam/ip-address-ranges", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(resourceName, "description", "managed by Terraform")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 65)
	start, end := testutil.AccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeConfig(cidr, start, end, fmt.Sprintf("tfacc-range-del-gone-%d", seed), "", false, false), Check: testutil.DeleteResourceOutOfBand("nautobot_ip_address_range.test", "ipam/ip-address-ranges")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_parallel(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidrs := []string{testutil.AccPrefixCIDR(seed, 67), testutil.AccPrefixCIDR(seed, 68), testutil.AccPrefixCIDR(seed, 69)}
	starts := make([]string, 3)
	ends := make([]string, 3)
	for i, cidr := range cidrs {
		starts[i], ends[i] = testutil.AccIPRangeBounds(cidr)
	}
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeParallelConfig(cidrs, starts, ends, seed), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("nautobot_ip_address_range.first", "id"),
			resource.TestCheckResourceAttrSet("nautobot_ip_address_range.second", "id"),
			resource.TestCheckResourceAttrSet("nautobot_ip_address_range.third", "id"),
		)},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_missingNamespaceAndParent(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 39)
	start, end := testutil.AccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeMissingNamespaceAndParentConfig(start, end, seed), ExpectError: regexp.MustCompile(`Invalid Attribute Combination`)},
	}})
}

func TestAccIPAddressRangeResource_provideParentAndNamespace(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 40)
	start, end := testutil.AccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeProvideParentAndNamespaceConfig(cidr, start, end, seed), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("nautobot_ip_address_range.test_parent_and_namespace", "id"), resource.TestCheckResourceAttrPair("nautobot_ip_address_range.test_parent_and_namespace", "parent_id", "nautobot_prefix.p", "id"), resource.TestCheckResourceAttrPair("nautobot_ip_address_range.test_parent_and_namespace", "namespace_id", "nautobot_prefix.p", "namespace_id"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "start_address", start), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "end_address", end), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "size", "3"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "ip_version", "4"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "count_as_utilized", "false"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "is_exclusive", "false"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_parent_and_namespace", "tags_ids.#", "0"), resource.TestCheckResourceAttrSet("nautobot_ip_address_range.test_parent_and_namespace", "created"),
		)},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccIPAddressRangeResource_provideOnlyNamespace(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 37)
	start, end := testutil.AccIPRangeBounds(cidr)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeProvideOnlyNamespaceConfig(cidr, start, end, seed), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet("nautobot_ip_address_range.test_only_namespace", "id"), resource.TestCheckResourceAttrPair("nautobot_ip_address_range.test_only_namespace", "namespace_id", "nautobot_prefix.p", "namespace_id"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "start_address", start), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "end_address", end), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "size", "3"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "ip_version", "4"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "count_as_utilized", "false"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "is_exclusive", "false"), resource.TestCheckResourceAttr("nautobot_ip_address_range.test_only_namespace", "tags_ids.#", "0"), resource.TestCheckResourceAttrSet("nautobot_ip_address_range.test_only_namespace", "created"),
		)},
		{Config: testutil.AccProviderConfig()},
	}})
}

func testAccIPAddressRangeChangeParentConfig(firstCIDR, secondCIDR, start, end, selectedParent string, seed int64) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "first" {
  name = "tfacc-range-first-%d"
}

resource "nautobot_namespace" "second" {
  name = "tfacc-range-second-%d"
}

resource "nautobot_prefix" "first" {
  prefix       = %q
  namespace_id = nautobot_namespace.first.id
  status       = %q
}

resource "nautobot_prefix" "second" {
  prefix       = %q
  namespace_id = nautobot_namespace.second.id
  status       = %q
}

resource "nautobot_ip_address_range" "changing_parent" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.%s.id
  status        = %q
  name          = "tfacc-range-change-parent-%d"
}
`, seed, seed, firstCIDR, testutil.Status, secondCIDR, testutil.Status,
		start, end, selectedParent, testutil.Status, seed)
}

func TestAccIPAddressRangeResource_changeParentAndDerivedNamespace(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	firstCIDR := testutil.AccPrefixCIDR(seed, 43)
	secondCIDR := testutil.AccPrefixCIDR(seed, 44)
	firstStart, firstEnd := testutil.AccIPRangeBounds(firstCIDR)
	secondStart, secondEnd := testutil.AccIPRangeBounds(secondCIDR)
	resourceName := "nautobot_ip_address_range.changing_parent"

	config := func(start, end, selectedParent string) string {
		return testAccIPAddressRangeChangeParentConfig(
			firstCIDR, secondCIDR, start, end, selectedParent, seed,
		)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(firstStart, firstEnd, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.first", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_namespace.first", "id"),
					resource.TestCheckResourceAttr(resourceName, "start_address", firstStart),
					resource.TestCheckResourceAttr(resourceName, "end_address", firstEnd),
				),
			},
			{
				Config: config(secondStart, secondEnd, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", "nautobot_prefix.second", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_namespace.second", "id"),
					resource.TestCheckResourceAttr(resourceName, "start_address", secondStart),
					resource.TestCheckResourceAttr(resourceName, "end_address", secondEnd),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
