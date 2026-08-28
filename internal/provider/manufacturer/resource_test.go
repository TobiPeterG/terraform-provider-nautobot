package manufacturer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const manufacturerResourceName = "nautobot_manufacturer.test"

func testAccManufacturerConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name = "%s"
}
`, name)
}

func testAccManufacturerConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%s"
  description = "created by terraform acceptance test"
}
`, name)
}

func testAccManufacturerConfigUpdated(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%s-updated"
  description = "updated by terraform acceptance test"
}
`, name)
}

func testAccManufacturerConfigParallel(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "m1" {
  name = "%s-1"
}

resource "nautobot_manufacturer" "m2" {
  name = "%s-2"
}

resource "nautobot_manufacturer" "m3" {
  name = "%s-3"
}
`, base, base, base)
}

func TestAccManufacturerResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "id"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "id"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-upd-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccManufacturerConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccManufacturerConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "updated by terraform acceptance test"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_drift(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-manufacturer-drift-%d", testutil.AccSeedForTest(t))
	config := testAccManufacturerConfigFull(name)
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID(manufacturerResourceName, &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "dcim/manufacturers", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr(manufacturerResourceName, "description", "created by terraform acceptance test")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccManufacturerResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-import-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
			},
			{
				ResourceName:      manufacturerResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-del-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-manufacturer-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccManufacturerConfigMinimal(name), Check: testutil.DeleteResourceOutOfBand(manufacturerResourceName, "dcim/manufacturers")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccManufacturerResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-manufacturer-par-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_manufacturer.m1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m1", "id"),

					resource.TestCheckResourceAttr("nautobot_manufacturer.m2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m2", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m2", "id"),

					resource.TestCheckResourceAttr("nautobot_manufacturer.m3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m3", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m3", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
