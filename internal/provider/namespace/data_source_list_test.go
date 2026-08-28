package namespace_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const namespacesDataSourceName = "data.nautobot_namespaces.test"

func testAccNamespacesDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "n1" {
  name = "%[1]s-1"
}

resource "nautobot_namespace" "n2" {
  name        = "%[1]s-2"
  description = "n2 created by terraform acceptance test"
}

resource "nautobot_namespace" "n3" {
  name        = "%[1]s-3"
  description = "n3 created by terraform acceptance test"
}

data "nautobot_namespaces" "test" {
  depends_on = [
    nautobot_namespace.n1,
    nautobot_namespace.n2,
    nautobot_namespace.n3,
  ]
}
`, base)
}

func TestAccNamespacesDataSource_list(t *testing.T) {
	t.Parallel()
	base := fmt.Sprintf("tfacc-ds-namespaces-%d", testutil.AccSeedForTest(t))
	n1, n2, n3 := base+"-1", base+"-2", base+"-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespacesDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(namespacesDataSourceName, "namespaces", 3),
					testutil.FindListIndexByAttr(namespacesDataSourceName, "namespaces", "name", n1),
					testutil.FindListIndexByAttr(namespacesDataSourceName, "namespaces", "name", n2),
					testutil.FindListIndexByAttr(namespacesDataSourceName, "namespaces", "name", n3),
					testutil.CheckNamespaceInListHasAttrs(namespacesDataSourceName, n1, map[string]string{
						"name": n1, "description": "", "location_id": "", "tenant_id": "",
					}),
					testutil.CheckNamespaceInListHasAttrs(namespacesDataSourceName, n2, map[string]string{
						"name": n2, "description": "n2 created by terraform acceptance test", "location_id": "", "tenant_id": "",
					}),
					testutil.CheckNamespaceInListHasAttrs(namespacesDataSourceName, n3, map[string]string{
						"name": n3, "description": "n3 created by terraform acceptance test", "location_id": "", "tenant_id": "",
					}),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
