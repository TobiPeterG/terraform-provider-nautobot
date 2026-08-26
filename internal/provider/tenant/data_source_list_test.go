package tenant_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const tenantsDataSourceName = "data.nautobot_tenants.test"

func testAccTenantsDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "t1" {
  name = "%[1]s-1"
}

resource "nautobot_tenant" "t2" {
  name        = "%[1]s-2"
  description = "t2 created by terraform acceptance test"
}

resource "nautobot_tenant" "t3" {
  name        = "%[1]s-3"
  description = "t3 created by terraform acceptance test"
  comments    = "t3 comment"
}

data "nautobot_tenants" "test" {
  depends_on = [
    nautobot_tenant.t1,
    nautobot_tenant.t2,
    nautobot_tenant.t3,
  ]
}
`, base)
}

func TestAccTenantsDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-tenants-%d", testutil.AccSeedForTest(t))

	t1 := base + "-1"
	t2 := base + "-2"
	t3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantsDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(tenantsDataSourceName, "tenants", 3),

					testutil.FindListIndexByAttr(tenantsDataSourceName, "tenants", "name", t1),
					testutil.FindListIndexByAttr(tenantsDataSourceName, "tenants", "name", t2),
					testutil.FindListIndexByAttr(tenantsDataSourceName, "tenants", "name", t3),

					testutil.CheckTenantInListHasAttrs(tenantsDataSourceName, t1, map[string]string{
						"name":        t1,
						"description": "",
						"comments":    "",
					}),
					testutil.CheckTenantInListHasAttrs(tenantsDataSourceName, t2, map[string]string{
						"name":        t2,
						"description": "t2 created by terraform acceptance test",
					}),
					testutil.CheckTenantInListHasAttrs(tenantsDataSourceName, t3, map[string]string{
						"name":        t3,
						"description": "t3 created by terraform acceptance test",
						"comments":    "t3 comment",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
