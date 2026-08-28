package tenant_group_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const tenantGroupsDataSourceName = "data.nautobot_tenant_groups.test"

func testAccTenantGroupsDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "g1" {
  name = "%[1]s-1"
}

resource "nautobot_tenant_group" "g2" {
  name        = "%[1]s-2"
  description = "g2 created by terraform acceptance test"
}

resource "nautobot_tenant_group" "g3" {
  name        = "%[1]s-3"
  description = "g3 created by terraform acceptance test"
}

data "nautobot_tenant_groups" "test" {
  depends_on = [
    nautobot_tenant_group.g1,
    nautobot_tenant_group.g2,
    nautobot_tenant_group.g3,
  ]
}
`, base)
}

func TestAccTenantGroupsDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-tgs-%d", testutil.AccSeedForTest(t))

	g1 := base + "-1"
	g2 := base + "-2"
	g3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupsDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(tenantGroupsDataSourceName, "tenant_groups", 3),

					testutil.FindListIndexByAttr(tenantGroupsDataSourceName, "tenant_groups", "name", g1),
					testutil.FindListIndexByAttr(tenantGroupsDataSourceName, "tenant_groups", "name", g2),
					testutil.FindListIndexByAttr(tenantGroupsDataSourceName, "tenant_groups", "name", g3),

					testutil.CheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g1, map[string]string{
						"name":        g1,
						"description": "",
					}),
					testutil.CheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g2, map[string]string{
						"name":        g2,
						"description": "g2 created by terraform acceptance test",
					}),
					testutil.CheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g3, map[string]string{
						"name":        g3,
						"description": "g3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
