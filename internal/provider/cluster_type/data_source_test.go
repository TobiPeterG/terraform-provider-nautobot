package cluster_type_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestAccClusterTypeDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-type-missing-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_cluster_type" "test" {
  name = %q
}
`, name),
			ExpectError: regexp.MustCompile(`Cluster type lookup failed`),
		}},
	})
}

const (
	clusterTypeDataSourceName = "data.nautobot_cluster_type.test"
)

func testAccClusterTypeDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name = "%s"
}

data "nautobot_cluster_type" "test" {
  id = nautobot_cluster_type.test.id
}
`, name)
}

func testAccClusterTypeDataSourceConfigWithDescription(name, description string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name        = "%s"
  description = "%s"
}

data "nautobot_cluster_type" "test" {
  name = nautobot_cluster_type.test.name
}
`, name, description)
}

func TestAccClusterTypeDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-cluster-type-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(
						clusterTypeDataSourceName, "id",
						"nautobot_cluster_type.test", "id",
					),

					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "description", ""),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "notes_url"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeDataSource_withDescription(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-cluster-type-desc-%d", testutil.AccSeedForTest(t))
	desc := "created by terraform acceptance test (data source)"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeDataSourceConfigWithDescription(name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "id"),

					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "description", desc),

					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "notes_url"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
