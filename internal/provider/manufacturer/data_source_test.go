package manufacturer_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	manufacturerDataSourceName = "data.nautobot_manufacturer.test"
)

func testAccManufacturerDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name = "%[1]s"
}

data "nautobot_manufacturer" "test" {
  id = nautobot_manufacturer.test.id
}
`, name)
}

func testAccManufacturerDataSourceConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
}

data "nautobot_manufacturer" "test" {
  name = nautobot_manufacturer.test.name
}
`, name)
}

func TestAccManufacturerDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						manufacturerDataSourceName, "id",
						"nautobot_manufacturer.test", "id",
					),

					resource.TestCheckResourceAttr(manufacturerDataSourceName, "description", ""),

					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						manufacturerDataSourceName, "id",
						"nautobot_manufacturer.test", "id",
					),

					resource.TestCheckResourceAttr(manufacturerDataSourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "last_updated"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-notfound-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_manufacturer" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Manufacturer lookup failed`),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
