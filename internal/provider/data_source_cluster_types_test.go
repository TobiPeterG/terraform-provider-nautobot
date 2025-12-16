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
	clusterTypesDataSourceName = "data.nautobot_cluster_types.test"
)

func testAccClusterTypesDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
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

func testCheckClusterTypesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		raw := rs.Primary.Attributes["cluster_types.#"]
		if raw == "" {
			return fmt.Errorf("%s: cluster_types.# is empty", dsAddr)
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse cluster_types.#=%q: %w", dsAddr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d cluster_types, got %d", dsAddr, min, n)
		}
		return nil
	}
}

func testFindClusterTypeIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		rawN := rs.Primary.Attributes["cluster_types.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse cluster_types.#=%q: %w", dsAddr, rawN, err)
		}
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("cluster_types.%d.name", i)
			if rs.Primary.Attributes[k] == wantName {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find cluster type name %q in cluster_types list", dsAddr, wantName)
	}
}

func testCheckClusterTypeInListHasAttrs(dsAddr, ctName string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes["cluster_types.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse cluster_types.#=%q: %w", dsAddr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			if rs.Primary.Attributes[fmt.Sprintf("cluster_types.%d.name", i)] == ctName {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find cluster type name %q in cluster_types list", dsAddr, ctName)
		}

		for field, expected := range want {
			k := fmt.Sprintf("cluster_types.%d.%s", idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}

		requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
		for _, f := range requiredComputed {
			k := fmt.Sprintf("cluster_types.%d.%s", idx, f)
			if rs.Primary.Attributes[k] == "" {
				return fmt.Errorf("%s: %s expected to be set, got empty", dsAddr, k)
			}
		}

		return nil
	}
}

func TestAccClusterTypesDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-cluster-types-%d", time.Now().Unix())

	ct1 := base + "-1"
	ct2 := base + "-2"
	ct3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypesDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckClusterTypesCountAtLeast(clusterTypesDataSourceName, 3),

					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct1),
					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct2),
					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct3),

					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct1, map[string]string{
						"name":        ct1,
						"description": "",
					}),

					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct2, map[string]string{
						"name":        ct2,
						"description": "ct2 created by terraform acceptance test",
					}),
					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct3, map[string]string{
						"name":        ct3,
						"description": "ct3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
