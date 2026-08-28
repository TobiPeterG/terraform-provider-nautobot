package manufacturer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	manufacturersDataSourceName = "data.nautobot_manufacturers.test"
)

func testAccManufacturersDataSourceConfig(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
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

func TestAccManufacturersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-manufacturers-%d", testutil.AccSeedForTest(t))

	m1 := base + "-1"
	m2 := base + "-2"
	m3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(manufacturersDataSourceName, "manufacturers", 3),

					testutil.FindListIndexByAttr(manufacturersDataSourceName, "manufacturers", "name", m1),
					testutil.FindListIndexByAttr(manufacturersDataSourceName, "manufacturers", "name", m2),
					testutil.FindListIndexByAttr(manufacturersDataSourceName, "manufacturers", "name", m3),

					testutil.CheckManufacturerInListHasAttrs(manufacturersDataSourceName, m1, map[string]string{
						"name":        m1,
						"description": "",
					}),
					testutil.CheckManufacturerInListHasAttrs(manufacturersDataSourceName, m2, map[string]string{
						"name":        m2,
						"description": "m2 created by terraform acceptance test",
					}),
					testutil.CheckManufacturerInListHasAttrs(manufacturersDataSourceName, m3, map[string]string{
						"name":        m3,
						"description": "m3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
