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
	_jsiiVlansDataSourceName = "data.nautobot_vlans.test"
)

func testAccVLANsDataSourceConfigBasic(base string, baseVid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "vlan1" {
  name   = "%[1]s-1"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_vlan" "vlan2" {
  name   = "%[1]s-2"
  vid    = %[2]d + 1
  status = "%[3]s"
}

resource "nautobot_vlan" "vlan3" {
  name   = "%[1]s-3"
  vid    = %[2]d + 2
  status = "%[3]s"
}

data "nautobot_vlans" "test" {
  depends_on = [
    nautobot_vlan.vlan1,
    nautobot_vlan.vlan2,
    nautobot_vlan.vlan3,
  ]
}
`, base, baseVid, status)
}

func testCheckVLANsCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		raw := rs.Primary.Attributes["vlans.#"]
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse vlans.#=%q: %w", dsAddr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d vlans, got %d", dsAddr, min, n)
		}
		return nil
	}
}

func testFindVLANIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		rawN := rs.Primary.Attributes["vlans.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse vlans.#=%q: %w", dsAddr, rawN, err)
		}
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("vlans.%d.name", i)
			if rs.Primary.Attributes[k] == wantName {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find VLAN name %q", dsAddr, wantName)
	}
}

func testCheckVLANInListHasAttrs(dsAddr, vlanName string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes["vlans.#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse vlans.#=%q: %w", dsAddr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			if rs.Primary.Attributes[fmt.Sprintf("vlans.%d.name", i)] == vlanName {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find VLAN %q", dsAddr, vlanName)
		}

		for field, expected := range want {
			k := fmt.Sprintf("vlans.%d.%s", idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}
		return nil
	}
}

func TestAccVLANsDataSource_basic(t *testing.T) {
	t.Parallel()

	seed := time.Now().Unix()
	base := fmt.Sprintf("tfacc-ds-vlans-basic-%d", seed)
	baseVid := testAccVLANVid(seed, 20)

	v1 := base + "-1"
	v2 := base + "-2"
	v3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANsDataSourceConfigBasic(base, baseVid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckVLANsCountAtLeast(_jsiiVlansDataSourceName, 3),

					testFindVLANIndexByName(_jsiiVlansDataSourceName, v1),
					testFindVLANIndexByName(_jsiiVlansDataSourceName, v2),
					testFindVLANIndexByName(_jsiiVlansDataSourceName, v3),

					testCheckVLANInListHasAttrs(_jsiiVlansDataSourceName, v1, map[string]string{
						"name":          v1,
						"status":        testStatus,
						"tenant_id":     "",
						"role_id":       "",
						"vlan_group_id": "",
						"prefix_count":  "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
