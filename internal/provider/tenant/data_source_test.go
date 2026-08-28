package tenant_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const tenantDataSourceName = "data.nautobot_tenant.test"

func testAccTenantDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name = "%[1]s"
}

data "nautobot_tenant" "test" {
  id = nautobot_tenant.test.id
}
`, name)
}

func testAccTenantDataSourceConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%[1]s-group"
}

resource "nautobot_tenant" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  comments    = "acceptance test comment"
  tenant_group_id = nautobot_tenant_group.test.id
}

data "nautobot_tenant" "test" {
  name = nautobot_tenant.test.name
}
`, name)
}

func TestAccTenantDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(tenantDataSourceName, "tenant_group_id", ""),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantDataSourceName, "id",
						"nautobot_tenant.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(tenantDataSourceName, "comments", "acceptance test comment"),
					resource.TestCheckResourceAttrPair(tenantDataSourceName, "tenant_group_id", "nautobot_tenant_group.test", "id"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantDataSourceName, "id",
						"nautobot_tenant.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccTenantDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-notfound-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_tenant" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Tenant lookup failed`),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
