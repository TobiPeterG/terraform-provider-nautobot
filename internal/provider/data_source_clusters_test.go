package provider

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	clustersDataSourceName = "data.nautobot_clusters.test"
)

func testAccClustersDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
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
`, base, testTenantID)
}

func testCheckClustersCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		raw := rs.Primary.Attributes["clusters.#"]
		if raw == "" {
			return fmt.Errorf("%s: clusters.# is empty", dsAddr)
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse clusters.#=%q: %w", dsAddr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d clusters, got %d", dsAddr, min, n)
		}
		return nil
	}
}

func testFindClusterIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		rawN := rs.Primary.Attributes["clusters.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse clusters.#=%q: %w", dsAddr, rawN, err)
		}
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("clusters.%d.name", i)
			if rs.Primary.Attributes[k] == wantName {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find cluster name %q in clusters list", dsAddr, wantName)
	}
}

func testCheckClusterInListHasAttrs(dsAddr, clName string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes["clusters.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse clusters.#=%q: %w", dsAddr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			if rs.Primary.Attributes[fmt.Sprintf("clusters.%d.name", i)] == clName {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find cluster name %q in clusters list", dsAddr, clName)
		}

		for field, expected := range want {
			k := fmt.Sprintf("clusters.%d.%s", idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}

		requiredComputed := []string{"id", "cluster_type_id", "created", "last_updated"}
		for _, f := range requiredComputed {
			k := fmt.Sprintf("clusters.%d.%s", idx, f)
			if rs.Primary.Attributes[k] == "" {
				return fmt.Errorf("%s: %s expected to be set, got empty", dsAddr, k)
			}
		}

		return nil
	}
}

func TestAccClustersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-clusters-%d", time.Now().Unix())

	cl1 := base + "-1"
	cl2 := base + "-2"
	cl3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckClustersCountAtLeast(clustersDataSourceName, 3),

					testFindClusterIndexByName(clustersDataSourceName, cl1),
					testFindClusterIndexByName(clustersDataSourceName, cl2),
					testFindClusterIndexByName(clustersDataSourceName, cl3),

					testCheckClusterInListHasAttrs(clustersDataSourceName, cl1, map[string]string{
						"name":             cl1,
						"tenant_id":        testTenantID,
						"comments":         "",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testCheckClusterInListHasAttrs(clustersDataSourceName, cl2, map[string]string{
						"name":             cl2,
						"tenant_id":        testTenantID,
						"comments":         "cl2 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testCheckClusterInListHasAttrs(clustersDataSourceName, cl3, map[string]string{
						"name":             cl3,
						"tenant_id":        testTenantID,
						"comments":         "cl3 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
