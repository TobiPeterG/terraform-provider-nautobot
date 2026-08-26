package namespace_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const namespaceDataSourceName = "data.nautobot_namespace.test"

func testAccNamespaceDataSourceConfigMinimal(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "test" {
  name = "%s"
}

data "nautobot_namespace" "test" {
  id = nautobot_namespace.test.id
}
`, name)
}

func testAccNamespaceDataSourceConfigFull(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "namespace_ds" {
  name = "%[1]s-tenant"
}

resource "nautobot_namespace" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  tenant_id   = nautobot_tenant.namespace_ds.id
}

data "nautobot_namespace" "test" {
  name = nautobot_namespace.test.name
}
`, name)
}

func TestAccNamespaceDataSource_minimal(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-min-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceDataSourceName, "name", name),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "id", namespaceResourceName, "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "location_id", ""),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "notes_url"),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccNamespaceDataSource_full(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-full-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceDataSourceName, "name", name),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "id", namespaceResourceName, "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "tenant_id", "nautobot_tenant.namespace_ds", "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "location_id", ""),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}

func TestAccNamespaceDataSource_notFound(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-notfound-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testutil.AccProviderConfig() + fmt.Sprintf(`
data "nautobot_namespace" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Namespace lookup failed`),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
