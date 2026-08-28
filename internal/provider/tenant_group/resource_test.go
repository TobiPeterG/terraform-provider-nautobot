package tenant_group_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const tenantGroupResourceName = "nautobot_tenant_group.test"

func testAccTenantGroupConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%s"
}
`, name)
}

func testAccTenantGroupConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name        = "%s"
  description = "created by terraform acceptance test"
}
`, name)
}

func testAccTenantGroupConfigUpdated(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name        = "%s-updated"
  description = "updated by terraform acceptance test"
}
`, name)
}

func testAccTenantGroupConfigNested(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "parent" {
  name        = "%s-parent"
  description = "parent group"
}

resource "nautobot_tenant_group" "test" {
  name        = "%s-child"
  description = "child group"
  parent_id   = nautobot_tenant_group.parent.id
}
`, name, name)
}

func testAccTenantGroupConfigChangingParent(name, selectedParent string) string {
	parentID := ""
	if selectedParent != "" {
		parentID = fmt.Sprintf("  parent_id = nautobot_tenant_group.%s.id\n", selectedParent)
	}
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "first" {
  name = "%[1]s-first"
}

resource "nautobot_tenant_group" "second" {
  name = "%[1]s-second"
}

resource "nautobot_tenant_group" "test" {
  name = "%[1]s-child"
%[2]s}
`, name, parentID)
}

func testAccTenantGroupConfigParallel(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "g1" {
  name = "%s-1"
}

resource "nautobot_tenant_group" "g2" {
  name        = "%s-2"
  description = "g2 description"
}

resource "nautobot_tenant_group" "g3" {
  name        = "%s-3"
  description = "g3 description"
}
`, base, base, base)
}

func TestAccTenantGroupResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "parent_id", ""),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-upd-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccTenantGroupConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTenantGroupConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "updated by terraform acceptance test"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_drift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-drift-%d", testutil.AccSeedForTest(t))
	driftedName := name + "-outside-terraform"
	var tenantGroupID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CaptureResourceID(tenantGroupResourceName, &tenantGroupID),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				PreConfig: func() {
					testutil.MutateResourceOutOfBand(t, tenantGroupID, "tenancy/tenant-groups", map[string]any{"name": driftedName, "description": "outside Terraform"})
				},
				Config:             testAccTenantGroupConfigFull(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTenantGroupConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccTenantGroupResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-import-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
			},
			{
				ResourceName:      tenantGroupResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-del-%d", testutil.AccSeedForTest(t))

	var tenantGroupID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckTenantGroupAbsent(&tenantGroupID),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
				Check:  testutil.CaptureResourceID(tenantGroupResourceName, &tenantGroupID),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
				Check:  testutil.DeleteResourceOutOfBand(tenantGroupResourceName, "tenancy/tenant-groups"),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccTenantGroupResource_nested(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-nested-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigNested(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_tenant_group.parent", "name", name+"-parent"),
					resource.TestCheckResourceAttr("nautobot_tenant_group.parent", "parent_id", ""),

					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name+"-child"),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "child group"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "parent_id"),
					resource.TestCheckResourceAttrPair(
						tenantGroupResourceName, "parent_id",
						"nautobot_tenant_group.parent", "id",
					),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_updateAndClearParent(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-parent-update-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigChangingParent(name, "first"),
				Check: resource.TestCheckResourceAttrPair(
					tenantGroupResourceName, "parent_id", "nautobot_tenant_group.first", "id",
				),
			},
			{
				Config: testAccTenantGroupConfigChangingParent(name, "second"),
				Check: resource.TestCheckResourceAttrPair(
					tenantGroupResourceName, "parent_id", "nautobot_tenant_group.second", "id",
				),
			},
			{
				Config: testAccTenantGroupConfigChangingParent(name, ""),
				Check:  resource.TestCheckResourceAttr(tenantGroupResourceName, "parent_id", ""),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccTenantGroupResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-tg-par-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_tenant_group.g1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_tenant_group.g1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_tenant_group.g1", "id"),

					resource.TestCheckResourceAttr("nautobot_tenant_group.g2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_tenant_group.g2", "description", "g2 description"),
					resource.TestCheckResourceAttrSet("nautobot_tenant_group.g2", "id"),

					resource.TestCheckResourceAttr("nautobot_tenant_group.g3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_tenant_group.g3", "description", "g3 description"),
					resource.TestCheckResourceAttrSet("nautobot_tenant_group.g3", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
