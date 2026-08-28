package cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	clustersDataSourceName = "data.nautobot_clusters.test"
)

func testAccClustersDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl1" {
  name            = "%[1]s-1"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
}

resource "nautobot_cluster" "cl2" {
  name            = "%[1]s-2"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
  comments        = "cl2 created by terraform acceptance test"
}

resource "nautobot_cluster" "cl3" {
  name            = "%[1]s-3"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
  comments        = "cl3 created by terraform acceptance test"
}

data "nautobot_clusters" "test" {
  depends_on = [
    nautobot_cluster.cl1,
    nautobot_cluster.cl2,
    nautobot_cluster.cl3,
  ]
}
`, base, "")
}

func TestAccClustersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-clusters-%d", testutil.AccSeedForTest(t))

	cl1 := base + "-1"
	cl2 := base + "-2"
	cl3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(clustersDataSourceName, "clusters", 3),

					testutil.FindListIndexByAttr(clustersDataSourceName, "clusters", "name", cl1),
					testutil.FindListIndexByAttr(clustersDataSourceName, "clusters", "name", cl2),
					testutil.FindListIndexByAttr(clustersDataSourceName, "clusters", "name", cl3),

					testutil.CheckClusterInListHasAttrs(clustersDataSourceName, cl1, map[string]string{
						"name":             cl1,
						"tenant_id":        "",
						"comments":         "",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testutil.CheckClusterInListHasAttrs(clustersDataSourceName, cl2, map[string]string{
						"name":             cl2,
						"tenant_id":        "",
						"comments":         "cl2 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testutil.CheckClusterInListHasAttrs(clustersDataSourceName, cl3, map[string]string{
						"name":             cl3,
						"tenant_id":        "",
						"comments":         "cl3 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
