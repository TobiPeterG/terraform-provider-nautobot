package cluster_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	clusterDataSourceName = "data.nautobot_cluster.test"
)

func testAccClusterDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "test" {
  name            = "%[1]s"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
}

data "nautobot_cluster" "test" {
  id = nautobot_cluster.test.id
}
`, name, "")
}

func testAccClusterDataSourceConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_tenant" "test" {
  name = "%[1]s-tenant"
}

resource "nautobot_cluster" "test" {
  name             = "%[1]s"
  cluster_type_id  = nautobot_cluster_type.ct.id
  tenant_id        = nautobot_tenant.test.id
  comments         = "created by terraform acceptance test"
  cluster_group_id = ""
  location_id      = ""
}

data "nautobot_cluster" "test" {
  name = nautobot_cluster.test.name
}
`, name)
}

func TestAccClusterDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-minimal-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "id",
						"nautobot_cluster.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "cluster_type_id",
						"nautobot_cluster.test", "cluster_type_id",
					),

					resource.TestCheckResourceAttr(clusterDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "location_id", ""),

					resource.TestCheckResourceAttr(clusterDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "notes_url"),
					resource.TestCheckResourceAttr(clusterDataSourceName, "device_count", "0"),
					resource.TestCheckResourceAttr(clusterDataSourceName, "virtual_machine_count", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "id",
						"nautobot_cluster.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "cluster_type_id",
						"nautobot_cluster.test", "cluster_type_id",
					),

					resource.TestCheckResourceAttr(clusterDataSourceName, "comments", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(clusterDataSourceName, "cluster_group_id", ""),
					resource.TestCheckResourceAttrPair(clusterDataSourceName, "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttr(clusterDataSourceName, "location_id", ""),

					resource.TestCheckResourceAttr(clusterDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-notfound-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_cluster" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Cluster lookup failed`),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
