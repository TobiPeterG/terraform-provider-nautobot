package ip_address_range_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	. "github.com/nautobot/terraform-provider-nautobot/internal/provider/ip_address_range"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
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
	return testutil.AccProviderConfig() + fmt.Sprintf(`
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
  parent_id      = nautobot_prefix.p1.id
  status        = %q
}
resource "nautobot_ip_address_range" "r2" {
  start_address     = %q
  end_address       = %q
  parent_id         = nautobot_prefix.p2.id
  status            = %q
  name              = %q
  description       = "created by terraform acceptance test"
  count_as_utilized = true
}
resource "nautobot_ip_address_range" "r3" {
  start_address = %q
  end_address   = %q
  parent_id     = nautobot_prefix.p3.id
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
`, cidr1, testutil.Status, cidr2, testutil.Status, cidr3, testutil.Status,
		start1, end1, testutil.Status,
		start2, end2, testutil.Status, fmt.Sprintf("tfacc-range-list-%d-2", seed),
		start3, end3, testutil.Status, fmt.Sprintf("tfacc-range-list-%d-3", seed))
}

func TestAccIPAddressRangesDataSource_list(t *testing.T) {
	t.Parallel()

	seed := testutil.AccSeedForTest(t)
	cidr1 := testutil.AccPrefixCIDR(seed, 31)
	cidr2 := testutil.AccPrefixCIDR(seed, 32)
	cidr3 := testutil.AccPrefixCIDR(seed, 33)
	start1, end1 := testutil.AccIPRangeBounds(cidr1)
	start2, end2 := testutil.AccIPRangeBounds(cidr2)
	start3, end3 := testutil.AccIPRangeBounds(cidr3)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAddressRangesDataSourceConfig(seed, cidr1, start1, end1, cidr2, start2, end2, cidr3, start3, end3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CountAtLeast(ipAddressRangesDataSourceName, "ip_address_ranges", 3),
					testutil.FindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start1),
					testutil.FindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start2),
					testutil.FindListIndexByAttr(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start3),
					testutil.CheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start1, map[string]string{
						"start_address":     start1,
						"end_address":       end1,
						"name":              "",
						"description":       "",
						"count_as_utilized": "false",
						"is_exclusive":      "false",
						"status":            testutil.Status,
						"ip_version":        "4",
						"size":              "3",
					}, nil),
					testutil.CheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start2, map[string]string{
						"name":              fmt.Sprintf("tfacc-range-list-%d-2", seed),
						"description":       "created by terraform acceptance test",
						"count_as_utilized": "true",
						"is_exclusive":      "false",
					}, nil),
					testutil.CheckListItemHasAttrs(ipAddressRangesDataSourceName, "ip_address_ranges", "start_address", start3, map[string]string{
						"name":         fmt.Sprintf("tfacc-range-list-%d-3", seed),
						"is_exclusive": "true",
					}, nil),
				),
			},
			{Config: testutil.AccProviderConfig()},
		},
	})
}
