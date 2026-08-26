package cluster_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func testAccClusterConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "test" {
  name            = "%s"
  cluster_type_id = nautobot_cluster_type.ct.id
}
`, name, name)
}

func testAccClusterConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_tenant" "test" {
  name = "%s-tenant"
}

resource "nautobot_cluster" "test" {
  name             = "%s-updated"
  cluster_type_id  = nautobot_cluster_type.ct.id
  tenant_id        = nautobot_tenant.test.id
  comments         = "updated comment"
}
`, name, name, name)
}

func testAccClusterConfigUpdated(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_tenant" "test" {
  name = "%s-tenant"
}

resource "nautobot_cluster" "test" {
  name            = "%s-updated"
  cluster_type_id = nautobot_cluster_type.ct.id
  comments        = "updated comment"
}
`, name, name, name)
}

func testAccClusterConfigUpdatedWithoutTenant(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "test" {
  name            = "%s-updated"
  cluster_type_id = nautobot_cluster_type.ct.id
  comments        = "updated comment"
}
`, name, name)
}

func testAccClusterConfigParallel(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl1" {
  name            = "%s-1"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_cluster" "cl2" {
  name            = "%s-2"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_cluster" "cl3" {
  name            = "%s-3"
  cluster_type_id = nautobot_cluster_type.ct.id
}
`, name,
		name,
		name,
		name,
	)
}

func TestAccClusterResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "cluster_type_id"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "cluster_group_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "location_id", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "created"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-upd-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", ""),
				),
			},
			{
				Config: testAccClusterConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", "updated comment"),
					resource.TestCheckResourceAttrPair("nautobot_cluster.test", "tenant_id", "nautobot_tenant.test", "id"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "cluster_group_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "location_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccClusterConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tenant_id", ""),
				),
			},
			{Config: testAccClusterConfigUpdatedWithoutTenant(name)},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_drift(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-cluster-drift-%d", testutil.AccSeedForTest(t))
	config := testAccClusterConfigFull(name)
	var id string
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: testutil.CaptureResourceID("nautobot_cluster.test", &id)},
		{PreConfig: func() {
			testutil.MutateResourceOutOfBand(t, id, "virtualization/clusters", map[string]any{"comments": "outside Terraform"})
		}, Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
		{Config: config, Check: resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", "updated comment")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccClusterResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-import-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
			},
			{
				ResourceName:      "nautobot_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-del-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_deleteAlreadyGone(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-cluster-del-gone-%d", testutil.AccSeedForTest(t))
	resource.Test(t, resource.TestCase{PreCheck: func() { testutil.AccPreCheck(t) }, ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories, AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}, Steps: []resource.TestStep{
		{Config: testAccClusterConfigMinimal(name), Check: testutil.DeleteResourceOutOfBand("nautobot_cluster.test", "virtualization/clusters")},
		{Config: testutil.AccProviderConfig()},
	}})
}

func TestAccClusterResource_parallel(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-par-%d", testutil.AccSeedForTest(t))

	resourceName1 := "nautobot_cluster.cl1"
	resourceName2 := "nautobot_cluster.cl2"
	resourceName3 := "nautobot_cluster.cl3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigParallel(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", name+"-1"),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName1, "comments", ""),
					resource.TestCheckResourceAttr(resourceName1, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "tags_ids.#", "0"),

					resource.TestCheckResourceAttr(resourceName2, "name", name+"-2"),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName2, "comments", ""),
					resource.TestCheckResourceAttr(resourceName2, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "tags_ids.#", "0"),

					resource.TestCheckResourceAttr(resourceName3, "name", name+"-3"),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName3, "comments", ""),
					resource.TestCheckResourceAttr(resourceName3, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "tags_ids.#", "0"),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}
