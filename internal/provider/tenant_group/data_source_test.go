package tenant_group_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const tenantGroupDataSourceName = "data.nautobot_tenant_group.test"

func testAccTenantGroupDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%[1]s"
}

data "nautobot_tenant_group" "test" {
  id = nautobot_tenant_group.test.id
}
`, name)
}

func testAccTenantGroupDataSourceConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "parent" {
  name = "%[1]s-parent"
}

resource "nautobot_tenant_group" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  parent_id   = nautobot_tenant_group.parent.id
}

data "nautobot_tenant_group" "test" {
  name = nautobot_tenant_group.test.name
}
`, name)
}

func TestAccTenantGroupDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantGroupDataSourceName, "id",
						"nautobot_tenant_group.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(tenantGroupDataSourceName, "parent_id", "nautobot_tenant_group.parent", "id"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantGroupDataSourceName, "id",
						"nautobot_tenant_group.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-notfound-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_tenant_group" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Tenant group lookup failed`),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
