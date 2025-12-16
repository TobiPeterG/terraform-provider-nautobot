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
	manufacturersDataSourceName = "data.nautobot_manufacturers.test"
)

func testAccManufacturersDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "m1" {
  name = "%[1]s-1"
}

resource "nautobot_manufacturer" "m2" {
  name        = "%[1]s-2"
  description = "m2 created by terraform acceptance test"
}

resource "nautobot_manufacturer" "m3" {
  name        = "%[1]s-3"
  description = "m3 created by terraform acceptance test"
}

data "nautobot_manufacturers" "test" {
  depends_on = [
    nautobot_manufacturer.m1,
    nautobot_manufacturer.m2,
    nautobot_manufacturer.m3,
  ]
}
`, base)
}

func testCheckManufacturersCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		raw := rs.Primary.Attributes["manufacturers.#"]
		if raw == "" {
			return fmt.Errorf("%s: manufacturers.# is empty", dsAddr)
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse manufacturers.#=%q: %w", dsAddr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d manufacturers, got %d", dsAddr, min, n)
		}
		return nil
	}
}

func testFindManufacturerIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		rawN := rs.Primary.Attributes["manufacturers.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse manufacturers.#=%q: %w", dsAddr, rawN, err)
		}
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("manufacturers.%d.name", i)
			if rs.Primary.Attributes[k] == wantName {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find manufacturer name %q in manufacturers list", dsAddr, wantName)
	}
}

func testCheckManufacturerInListHasAttrs(dsAddr, mName string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes["manufacturers.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse manufacturers.#=%q: %w", dsAddr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			if rs.Primary.Attributes[fmt.Sprintf("manufacturers.%d.name", i)] == mName {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find manufacturer name %q in manufacturers list", dsAddr, mName)
		}

		for field, expected := range want {
			k := fmt.Sprintf("manufacturers.%d.%s", idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}

		requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
		for _, f := range requiredComputed {
			k := fmt.Sprintf("manufacturers.%d.%s", idx, f)
			if rs.Primary.Attributes[k] == "" {
				return fmt.Errorf("%s: %s expected to be set, got empty", dsAddr, k)
			}
		}

		return nil
	}
}

func TestAccManufacturersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-manufacturers-%d", time.Now().Unix())

	m1 := base + "-1"
	m2 := base + "-2"
	m3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckManufacturersCountAtLeast(manufacturersDataSourceName, 3),

					testFindManufacturerIndexByName(manufacturersDataSourceName, m1),
					testFindManufacturerIndexByName(manufacturersDataSourceName, m2),
					testFindManufacturerIndexByName(manufacturersDataSourceName, m3),

					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m1, map[string]string{
						"name":        m1,
						"description": "",
					}),
					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m2, map[string]string{
						"name":        m2,
						"description": "m2 created by terraform acceptance test",
					}),
					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m3, map[string]string{
						"name":        m3,
						"description": "m3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
