package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantResourceName = "nautobot_tenant.test"

func testAccTenantConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name = "%s"
}
`, name)
}

func testAccTenantConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name        = "%s"
  description = "created by terraform acceptance test"
  comments    = "acceptance test comment"
}
`, name)
}

func testAccTenantConfigUpdated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name        = "%s-updated"
  description = "updated by terraform acceptance test"
  comments    = "updated comment"
}
`, name)
}

func testAccTenantConfigWithGroup(name, groupName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%s-group"
}

resource "nautobot_tenant" "test" {
  name         = "%s"
  description  = "tenant with group"
  tenant_group = nautobot_tenant_group.test.id
}
`, groupName, name)
}

func testAccTenantConfigParallel(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "t1" {
  name = "%s-1"
}

resource "nautobot_tenant" "t2" {
  name        = "%s-2"
  description = "t2 description"
}

resource "nautobot_tenant" "t3" {
  name        = "%s-3"
  description = "t3 description"
  comments    = "t3 comment"
}
`, base, base, base)
}

func TestAccTenantResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantResourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantResourceName, "comments", ""),
					resource.TestCheckResourceAttr(tenantResourceName, "tenant_group", ""),
					resource.TestCheckResourceAttrSet(tenantResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(tenantResourceName, "comments", "acceptance test comment"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "id"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccTenantConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccTenantConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(tenantResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(tenantResourceName, "comments", "updated comment"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigMinimal(name),
			},
			{
				ResourceName:      tenantResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_withTenantGroup(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-tenant-wg-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigWithGroup(name, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(tenantResourceName, "name", name),
					resource.TestCheckResourceAttr(tenantResourceName, "description", "tenant with group"),
					resource.TestCheckResourceAttrSet(tenantResourceName, "tenant_group"),
					resource.TestCheckResourceAttrPair(
						tenantResourceName, "tenant_group",
						"nautobot_tenant_group.test", "id",
					),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-tenant-par-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_tenant.t1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_tenant.t1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_tenant.t1", "id"),

					resource.TestCheckResourceAttr("nautobot_tenant.t2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_tenant.t2", "description", "t2 description"),
					resource.TestCheckResourceAttrSet("nautobot_tenant.t2", "id"),

					resource.TestCheckResourceAttr("nautobot_tenant.t3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_tenant.t3", "description", "t3 description"),
					resource.TestCheckResourceAttr("nautobot_tenant.t3", "comments", "t3 comment"),
					resource.TestCheckResourceAttrSet("nautobot_tenant.t3", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
