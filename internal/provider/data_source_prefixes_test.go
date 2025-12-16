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
	prefixesDataSourceName = "data.nautobot_prefixes.test"
)

func testAccPrefixesDataSourceConfigBasic(base string) string {
	p1 := fmt.Sprintf("10.251.%d.0/24", int(time.Now().Unix()%200)+1)
	p2 := fmt.Sprintf("10.252.%d.0/24", int(time.Now().Unix()%200)+1)
	p3 := fmt.Sprintf("10.253.%d.0/24", int(time.Now().Unix()%200)+1)

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p1" {
  prefix  = "%[2]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[3]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[4]s"
  status  = "%[5]s"
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, testStatus)
}

func testAccPrefixesDataSourceConfigFull(base string) string {
	p1 := fmt.Sprintf("10.241.%d.0/24", int(time.Now().Unix()%200)+1)
	p2 := fmt.Sprintf("10.242.%d.0/24", int(time.Now().Unix()%200)+1)
	p3 := fmt.Sprintf("10.243.%d.0/24", int(time.Now().Unix()%200)+1)

	return testAccProviderConfig() + fmt.Sprintf(`
# minimal prefix
resource "nautobot_prefix" "p1" {
  prefix = "%[2]s"
  status = "%[6]s"
}

# full prefix A
resource "nautobot_prefix" "p2" {
  prefix      = "%[3]s"
  status      = "%[6]s"
  description = "p2 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = "%[7]s"
}

# full prefix B (different values)
resource "nautobot_prefix" "p3" {
  prefix      = "%[4]s"
  status      = "%[6]s"
  description = "p3 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = "%[7]s"
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, testTenantID, testStatus, testVLAN)
}

func testCheckPrefixesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		raw := rs.Primary.Attributes["prefixes.#"]
		if raw == "" {
			return fmt.Errorf("%s: prefixes.# is empty", dsAddr)
		}
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse prefixes.#=%q: %w", dsAddr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d prefixes, got %d", dsAddr, min, n)
		}
		return nil
	}
}

func testFindPrefixIndexByCIDR(dsAddr, wantCIDR string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		rawN := rs.Primary.Attributes["prefixes.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse prefixes.#=%q: %w", dsAddr, rawN, err)
		}
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("prefixes.%d.prefix", i)
			if rs.Primary.Attributes[k] == wantCIDR {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find prefix %q in prefixes list", dsAddr, wantCIDR)
	}
}

func testCheckPrefixInListHasAttrs(dsAddr, cidr string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes["prefixes.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse prefixes.#=%q: %w", dsAddr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			if rs.Primary.Attributes[fmt.Sprintf("prefixes.%d.prefix", i)] == cidr {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find prefix %q in prefixes list", dsAddr, cidr)
		}

		for field, expected := range want {
			k := fmt.Sprintf("prefixes.%d.%s", idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}
		return nil
	}
}

func TestAccPrefixesDataSource_basic(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-prefixes-basic-%d", time.Now().Unix())

	p1 := fmt.Sprintf("10.251.%d.0/24", int(time.Now().Unix()%200)+1)
	p2 := fmt.Sprintf("10.252.%d.0/24", int(time.Now().Unix()%200)+1)
	p3 := fmt.Sprintf("10.253.%d.0/24", int(time.Now().Unix()%200)+1)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigBasic(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckPrefixesCountAtLeast(prefixesDataSourceName, 3),

					testFindPrefixIndexByCIDR(prefixesDataSourceName, p1),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p2),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p3),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":         p1,
						"description":    "",
						"tenant_id":      "",
						"role_id":        "",
						"parent_id":      "",
						"rir_id":         "",
						"namespace_id":   "",
						"vlan_id":        "",
						"date_allocated": "",
						"tags_ids.#":     "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixesDataSource_full(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-prefixes-full-%d", time.Now().Unix())

	p1 := fmt.Sprintf("10.241.%d.0/24", int(time.Now().Unix()%200)+1)
	p2 := fmt.Sprintf("10.242.%d.0/24", int(time.Now().Unix()%200)+1)
	p3 := fmt.Sprintf("10.243.%d.0/24", int(time.Now().Unix()%200)+1)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigFull(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckPrefixesCountAtLeast(prefixesDataSourceName, 3),

					testFindPrefixIndexByCIDR(prefixesDataSourceName, p1),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p2),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p3),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":         p1,
						"description":    "",
						"tenant_id":      "",
						"vlan_id":        "",
						"date_allocated": "",
						"tags_ids.#":     "0",
					}),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p2, map[string]string{
						"prefix":         p2,
						"description":    "p2 created by terraform acceptance test",
						"tenant_id":      testTenantID,
						"vlan_id":        testVLAN,
						"date_allocated": "",
						"tags_ids.#":     "0",
					}),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p3, map[string]string{
						"prefix":         p3,
						"description":    "p3 created by terraform acceptance test",
						"tenant_id":      testTenantID,
						"vlan_id":        testVLAN,
						"date_allocated": "",
						"tags_ids.#":     "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
