package prefix_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/nautobot/terraform-provider-nautobot/internal/provider/prefix"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/shared"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	prefixResourceName = "nautobot_prefix.test"
)

func TestPrefixDerivedAttributesHaveNoPlanModifiers(t *testing.T) {
	t.Parallel()
	testutil.AssertStringAttributesHaveNoPlanModifiers(t, NewPrefixResource(), "parent_id", "rir_id", "namespace_id", "network", "broadcast")
}

func TestPrefixDateAllocatedIsManagedOptionalAttribute(t *testing.T) {
	t.Parallel()

	var response frameworkresource.SchemaResponse
	NewPrefixResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected resource schema diagnostics: %v", response.Diagnostics)
	}
	attribute, ok := response.Schema.Attributes["date_allocated"].(rschema.StringAttribute)
	if !ok {
		t.Fatal("date_allocated is not a string attribute")
	}
	if !attribute.Optional || attribute.Computed {
		t.Fatalf("date_allocated must be Optional and not Computed, got Optional=%t Computed=%t", attribute.Optional, attribute.Computed)
	}
	if _, ok := attribute.CustomType.(shared.RFC3339InstantType); !ok {
		t.Fatalf("date_allocated custom type = %T, want shared.RFC3339InstantType", attribute.CustomType)
	}
}

func TestAccPrefixResource_rejectsEmptyDateAllocated(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	config := testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "test" {
  name = "tfacc-prefix-empty-date-%d"
}

resource "nautobot_prefix" "test" {
  prefix         = %q
  namespace_id   = nautobot_namespace.test.id
  status         = %q
  date_allocated = ""
}
`, seed, testutil.AccPrefixCIDR(seed, 74), testutil.Status)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`Invalid RFC3339 timestamp`),
		}},
	})
}

func TestAccPrefixResource_dateAllocatedNormalization(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	resourceName := "nautobot_prefix.test"
	configuredDate := "2026-01-02T04:04:05.123400+01:00"
	config := testAccPrefixIdentityAndDateConfig(testutil.AccPrefixCIDR(seed, 73), configuredDate, seed)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(resourceName, "date_allocated", configuredDate),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func testAccPrefixConfigMinimal(name string, vid int, cidr string) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigFull(name string, vid int, cidr string) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_tenant" "test" {
  name = "%[1]s-tenant"
}

resource "nautobot_prefix" "test" {
  prefix         = "%[4]s"
  status         = "%[3]s"
  vlan_id        = nautobot_vlan.v.id
  tenant_id      = nautobot_tenant.test.id
  description    = "created by terraform acceptance test"
  date_allocated = "2026-01-02T03:04:05Z"
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigUpdated(name string, vid int, cidr string) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_tenant" "test" {
  name = "%[1]s-tenant"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "updated by terraform acceptance test"
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigUpdatedWithoutTenant(name string, vid int, cidr string) string {
	status := testutil.Status

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "updated by terraform acceptance test"
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigParallel(baseName string, baseVid int, seed int64) string {
	status := testutil.Status
	c1 := testutil.AccPrefixCIDR(seed, 8)
	c2 := testutil.AccPrefixCIDR(seed, 9)
	c3 := testutil.AccPrefixCIDR(seed, 10)

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p1" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[5]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[6]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, baseName, baseVid, status, c1, c2, c3)
}

func TestAccPrefixResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-minimal-%d", seed)
	vid := testutil.AccVLANVID(seed, 14)
	cidr := testutil.AccPrefixCIDR(seed, 11)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testutil.Status),

					resource.TestCheckResourceAttrSet(prefixResourceName, "id"),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr(prefixResourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "rir_id", ""),
					resource.TestCheckNoResourceAttr(prefixResourceName, "date_allocated"),

					resource.TestCheckResourceAttr(prefixResourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(prefixResourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "ip_version"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "namespace_id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_full(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-full-%d", seed)
	vid := testutil.AccVLANVID(seed, 15)
	cidr := testutil.AccPrefixCIDR(seed, 12)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testutil.Status),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "date_allocated", "2026-01-02T03:04:05Z"),
					resource.TestCheckResourceAttrPair(prefixResourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_namespace(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr := testutil.AccPrefixCIDR(seed, 36)
	config := testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "test" {
  name = "tf-acc-prefix-namespace-%d"
}

resource "nautobot_prefix" "test" {
  prefix       = %q
  status       = %q
  namespace_id = nautobot_namespace.test.id
}
`, seed, cidr, testutil.Status)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttrPair(prefixResourceName, "namespace_id", "nautobot_namespace.test", "id"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccPrefixResource_update(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-update-%d", seed)
	vid := testutil.AccVLANVID(seed, 16)
	cidr := testutil.AccPrefixCIDR(seed, 13)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccPrefixConfigUpdated(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccPrefixConfigUpdated(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
				),
			},
			{Config: testAccPrefixConfigUpdatedWithoutTenant(name, vid, cidr)},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_drift(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-prefix-drift-%d", seed)
	config := testAccPrefixConfigFull(name, testutil.AccVLANVID(seed, 70), testutil.AccPrefixCIDR(seed, 70))
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(prefixResourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "ipam/prefixes", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func testAccPrefixIdentityAndDateConfig(cidr, dateAllocated string, seed int64) string {
	dateAllocatedConfig := ""
	if dateAllocated != "" {
		dateAllocatedConfig = fmt.Sprintf("date_allocated = %q", dateAllocated)
	}

	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "identity" {
  name = "tf-acc-prefix-identity-%d"
}

resource "nautobot_prefix" "test" {
  prefix         = %q
  namespace_id   = nautobot_namespace.identity.id
  status         = %q
  %s
}
`, seed, cidr, testutil.Status, dateAllocatedConfig)
}

func TestAccPrefixResource_updateIdentityAndDate(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	firstCIDR := testutil.AccPrefixCIDR(seed, 45)
	secondCIDR := testutil.AccPrefixCIDR(seed, 46)
	resourceName := prefixResourceName
	firstDate := "2026-02-03T04:05:06Z"
	secondDate := "2026-03-04T05:06:07Z"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixIdentityAndDateConfig(firstCIDR, firstDate, seed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix", firstCIDR),
					resource.TestCheckResourceAttr(resourceName, "date_allocated", firstDate),
					resource.TestCheckResourceAttrPair(resourceName, "namespace_id", "nautobot_namespace.identity", "id"),
				),
			},
			{
				Config: testAccPrefixIdentityAndDateConfig(secondCIDR, secondDate, seed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix", secondCIDR),
					resource.TestCheckResourceAttr(resourceName, "date_allocated", secondDate),
					resource.TestCheckResourceAttrSet(resourceName, "network"),
					resource.TestCheckResourceAttrSet(resourceName, "broadcast"),
				),
			},
			{
				Config: testAccPrefixIdentityAndDateConfig(secondCIDR, "", seed),
				Check:  resource.TestCheckNoResourceAttr(resourceName, "date_allocated"),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccPrefixResource_import(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-import-%d", seed)
	vid := testutil.AccVLANVID(seed, 17)
	cidr := testutil.AccPrefixCIDR(seed, 14)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				ResourceName:      prefixResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_delete(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-delete-%d", seed)
	vid := testutil.AccVLANVID(seed, 18)
	cidr := testutil.AccPrefixCIDR(seed, 15)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	seed := testutil.AccSeedForTest(t)
	name := fmt.Sprintf("tfacc-prefix-del-gone-%d", seed)
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccPrefixConfigMinimal(name, testutil.AccVLANVID(seed, 62), testutil.AccPrefixCIDR(seed, 62)), Check: testutil.DeleteResourceOutOfBand(prefixResourceName, "ipam/prefixes")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccPrefixResource_parallel(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	baseName := fmt.Sprintf("tf-acc-prefix-parallel-%d", seed)
	baseVid := testutil.AccVLANVID(seed, 19)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigParallel(baseName, baseVid, seed),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("nautobot_prefix.p1", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p2", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p3", "id"),

					resource.TestCheckResourceAttrPair("nautobot_prefix.p1", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p2", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p3", "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "status", testutil.Status),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "status", testutil.Status),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "status", testutil.Status),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "description", ""),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
