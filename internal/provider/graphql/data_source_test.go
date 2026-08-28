package graphql_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

const (
	graphQLDataSourceName = "data.nautobot_graphql.test"
)

func testAccGraphQLDataSourceConfigManufacturers(name string) string {
	return testutil.AccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name = %q
}

data "nautobot_graphql" "test" {
  depends_on = [nautobot_manufacturer.test]
  query = <<-GQL
query {
  manufacturers {
    name
  }
}
GQL
}
`, name)
}

func testCheckGraphQLDataContains(dsAddr, substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		got := rs.Primary.Attributes["data"]
		if got == "" {
			return fmt.Errorf("%s: data is empty", dsAddr)
		}
		if !strings.Contains(got, substr) {
			return fmt.Errorf("%s: expected data to contain %q, got %q", dsAddr, substr, got)
		}
		return nil
	}
}

func TestAccGraphQLDataSource_returnsCreatedManufacturer(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-graphql-manufacturer-%d", testutil.AccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLDataSourceConfigManufacturers(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(graphQLDataSourceName, "query"),
					resource.TestCheckResourceAttrSet(graphQLDataSourceName, "data"),
					testCheckGraphQLDataContains(graphQLDataSourceName, name),
				),
			},
			{
				Config: testutil.AccProviderConfig(),
			},
		},
	})
}

func TestAccGraphQLDataSource_reportsGraphQLErrors(t *testing.T) {
	t.Parallel()

	config := testutil.AccProviderConfig() + `
data "nautobot_graphql" "test" {
  query = "query { this_field_does_not_exist }"
}
`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		Steps: []resource.TestStep{{
			Config:      config,
			ExpectError: regexp.MustCompile(`Failed to execute GraphQL query`),
		}},
	})
}
