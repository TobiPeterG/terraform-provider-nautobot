package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantGroupResourceName = "nautobot_tenant_group.test"

func testAccTenantGroupConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%s"
}
`, name)
}

func testAccTenantGroupConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name        = "%s"
  description = "created by terraform acceptance test"
}
`, name)
}

func testAccTenantGroupConfigUpdated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name        = "%s-updated"
  description = "updated by terraform acceptance test"
}
`, name)
}

func testAccTenantGroupConfigNested(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "parent" {
  name        = "%s-parent"
  description = "parent group"
}

resource "nautobot_tenant_group" "test" {
  name        = "%s-child"
  description = "child group"
  parent      = nautobot_tenant_group.parent.id
}
`, name, name)
}

func testAccTenantGroupConfigParallel(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
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

	name := fmt.Sprintf("tfacc-tg-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "parent", ""),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_nested(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tg-nested-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupConfigNested(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_tenant_group.parent", "name", name+"-parent"),
					resource.TestCheckResourceAttr("nautobot_tenant_group.parent", "parent", ""),

					resource.TestCheckResourceAttr(tenantGroupResourceName, "name", name+"-child"),
					resource.TestCheckResourceAttr(tenantGroupResourceName, "description", "child group"),
					resource.TestCheckResourceAttrSet(tenantGroupResourceName, "parent"),
					resource.TestCheckResourceAttrPair(
						tenantGroupResourceName, "parent",
						"nautobot_tenant_group.parent", "id",
					),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-tg-par-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
				Config: testAccProviderConfig(),
			},
		},
	})
}
