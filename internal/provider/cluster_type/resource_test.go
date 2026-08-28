package cluster_type_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func testAccClusterTypeConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name = "%s"
}
`, name)
}

func testAccClusterTypeConfigWithDescription(name, description string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name        = "%s"
  description = "%s"
}
`, name, description)
}

func testAccClusterTypeConfigParallel(base string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct1" {
  name = "%s-1"
}

resource "nautobot_cluster_type" "ct2" {
  name = "%s-2"
}

resource "nautobot_cluster_type" "ct3" {
  name = "%s-3"
}
`, base, base, base)
}

func TestAccClusterTypeResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "id"),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "created"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-upd-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", ""),
				),
			},
			{
				Config: testAccClusterTypeConfigWithDescription(name, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", "updated description"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_drift(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-cluster-type-drift-%d", testutil.AccSeedForTest(t))
	config := testAccClusterTypeConfigWithDescription(name, "managed by Terraform")
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID("nautobot_cluster_type.test", &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "virtualization/cluster-types", map[string]any{"description": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", "managed by Terraform")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccClusterTypeResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-import-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
			},
			{
				ResourceName:      "nautobot_cluster_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-del-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-cluster-type-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccClusterTypeConfigMinimal(name), Check: testutil.DeleteResourceOutOfBand("nautobot_cluster_type.test", "virtualization/cluster-types")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccClusterTypeResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-cluster-type-par-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct1", "id"),

					resource.TestCheckResourceAttr("nautobot_cluster_type.ct2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct2", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct2", "id"),

					resource.TestCheckResourceAttr("nautobot_cluster_type.ct3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct3", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct3", "id"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
