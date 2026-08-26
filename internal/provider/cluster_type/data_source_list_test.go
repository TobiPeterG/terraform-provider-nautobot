package cluster_type_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	clusterTypesDataSourceName = "data.nautobot_cluster_types.test"
)

func testAccClusterTypesDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct1" {
  name = "%[1]s-1"
}

resource "nautobot_cluster_type" "ct2" {
  name        = "%[1]s-2"
  description = "ct2 created by terraform acceptance test"
}

resource "nautobot_cluster_type" "ct3" {
  name        = "%[1]s-3"
  description = "ct3 created by terraform acceptance test"
}

data "nautobot_cluster_types" "test" {
  depends_on = [
    nautobot_cluster_type.ct1,
    nautobot_cluster_type.ct2,
    nautobot_cluster_type.ct3,
  ]
}
`, base)
}

func TestAccClusterTypesDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-cluster-types-%d", testutil.AccSeedForTest(t))

	ct1 := base + "-1"
	ct2 := base + "-2"
	ct3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypesDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(clusterTypesDataSourceName, "cluster_types", 3),

					testutil.FindListIndexByAttr(clusterTypesDataSourceName, "cluster_types", "name", ct1),
					testutil.FindListIndexByAttr(clusterTypesDataSourceName, "cluster_types", "name", ct2),
					testutil.FindListIndexByAttr(clusterTypesDataSourceName, "cluster_types", "name", ct3),

					testutil.CheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct1, map[string]string{
						"name":        ct1,
						"description": "",
					}),

					testutil.CheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct2, map[string]string{
						"name":        ct2,
						"description": "ct2 created by terraform acceptance test",
					}),
					testutil.CheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct3, map[string]string{
						"name":        ct3,
						"description": "ct3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
