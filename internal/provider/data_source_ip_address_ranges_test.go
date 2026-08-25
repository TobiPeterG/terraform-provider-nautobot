package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const ipAddressRangesDataSourceName = "data.nautobot_ip_address_ranges.test"

func TestIPAddressRangesDataSourceSchemaHasNoSelectorDescriptions(t *testing.T) {
	t.Parallel()

	var resp datasource.SchemaResponse
	(&IPAddressRangesDataSource{}).Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}

	ranges, ok := resp.Schema.Attributes["ip_address_ranges"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("ip_address_ranges is not a list nested attribute")
	}
	for _, name := range []string{"id", "start_address", "end_address", "namespace_id"} {
		attribute, ok := ranges.NestedObject.Attributes[name].(dsschema.StringAttribute)
		if !ok {
			t.Fatalf("%s is not a string attribute", name)
		}
		if attribute.Optional || !attribute.Computed {
			t.Errorf("plural item attribute %s must be computed-only", name)
		}
		if strings.Contains(attribute.Description, "Provide") {
			t.Errorf("plural item attribute %s contains singular selector guidance: %q", name, attribute.Description)
		}
	}
}

func testAccIPAddressRangesDataSourceConfig(seed int64, cidr1, start1, end1, cidr2, start2, end2, cidr3, start3, end3 string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p1" {
  prefix = %q
  status = %q
}
resource "nautobot_prefix" "p2" {
  prefix = %q
  status = %q
}
resource "nautobot_prefix" "p3" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "r1" {
  start_address = %q
  end_address   = %q
  status        = %q
}
resource "nautobot_ip_address_range" "r2" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = %q
  description       = "created by terraform acceptance test"
  count_as_utilized = true
}
resource "nautobot_ip_address_range" "r3" {
  start_address = %q
  end_address   = %q
  status        = %q
  name          = %q
  is_exclusive  = true
}

data "nautobot_ip_address_ranges" "test" {
  depends_on = [
    nautobot_ip_address_range.r1,
    nautobot_ip_address_range.r2,
    nautobot_ip_address_range.r3,
  ]
}
`, cidr1, testStatus, cidr2, testStatus, cidr3, testStatus,
		start1, end1, testStatus,
		start2, end2, testStatus, fmt.Sprintf("tfacc-range-list-%d-2", seed),
		start3, end3, testStatus, fmt.Sprintf("tfacc-range-list-%d-3", seed))
}

func TestAccIPAddressRangesDataSource_list(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	cidr1 := testAccPrefixCIDR(seed, 31)
	cidr2 := testAccPrefixCIDR(seed, 32)
	cidr3 := testAccPrefixCIDR(seed, 33)
	start1, end1 := testAccIPRangeBounds(cidr1)
	start2, end2 := testAccIPRangeBounds(cidr2)
	start3, end3 := testAccIPRangeBounds(cidr3)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressRangesDataSourceConfig(seed, cidr1, start1, end1, cidr2, start2, end2, cidr3, start3, end3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCountAtLeast(ipAddressRangesDataSourceName, "ip_address_ranges", 3),
					testFindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start1),
					testFindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start2),
					testFindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start3),
					testCheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start1, map[string]string{
						"start_address":     start1,
						"end_address":       end1,
						"name":              "",
						"description":       "",
						"count_as_utilized": "false",
						"is_exclusive":      "false",
						"status":            testStatus,
						"ip_version":        "4",
						"size":              "3",
					}, nil),
					testCheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start2, map[string]string{
						"name":              fmt.Sprintf("tfacc-range-list-%d-2", seed),
						"description":       "created by terraform acceptance test",
						"count_as_utilized": "true",
						"is_exclusive":      "false",
					}, nil),
					testCheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start3, map[string]string{
						"name":         fmt.Sprintf("tfacc-range-list-%d-3", seed),
						"is_exclusive": "true",
					}, nil),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}
