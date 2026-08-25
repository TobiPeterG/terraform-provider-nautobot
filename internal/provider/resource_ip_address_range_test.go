package provider

import (
	"context"
	"fmt"
	"net"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestIPAddressRangeResourceSchema(t *testing.T) {
	t.Parallel()
	var resp frameworkresource.SchemaResponse
	(&IPAddressRangeResource{}).Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", resp.Diagnostics)
	}
	for _, name := range []string{"start_address", "end_address", "namespace_id", "parent_id", "name", "description", "count_as_utilized", "is_exclusive", "status", "role_id", "tenant_id", "tags_ids", "size", "ip_version"} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("IP address range resource schema is missing %q", name)
		}
	}
	if _, ok := resp.Schema.Attributes["last_updated"]; ok {
		t.Error("IP address range resource must not expose last_updated; managed resources omit it")
	}
}

func testAccIPRangeBounds(cidr string) (string, string) {
	ip, _, _ := net.ParseCIDR(cidr)
	ip = ip.To4()
	return fmt.Sprintf("%d.%d.%d.10", ip[0], ip[1], ip[2]), fmt.Sprintf("%d.%d.%d.12", ip[0], ip[1], ip[2])
}

func testAccIPAddressRangeConfig(seed int64, cidr, start, end, name, description string, utilized bool) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p" {
  prefix = %q
  status = %q
}

resource "nautobot_ip_address_range" "test" {
  start_address     = %q
  end_address       = %q
  status            = %q
  name              = %q
  description       = %q
  count_as_utilized = %t
  is_exclusive      = false
}
`, cidr, testStatus, start, end, testStatus, name, description, utilized)
}

func TestAccIPAddressRangeResource_lifecycle(t *testing.T) {
	t.Parallel()
	seed := testAccSeedForTest(t)
	cidr := testAccPrefixCIDR(seed, 24)
	start, end := testAccIPRangeBounds(cidr)
	addr := "nautobot_ip_address_range.test"
	resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) }, ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: testAccIPAddressRangeConfig(seed, cidr, start, end, fmt.Sprintf("tfacc-range-%d", seed), "created by terraform", false), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrSet(addr, "id"), resource.TestCheckResourceAttrPair(addr, "parent_id", "nautobot_prefix.p", "id"), resource.TestCheckResourceAttr(addr, "start_address", start), resource.TestCheckResourceAttr(addr, "end_address", end), resource.TestCheckResourceAttr(addr, "size", "3"), resource.TestCheckResourceAttr(addr, "ip_version", "4"), resource.TestCheckResourceAttr(addr, "count_as_utilized", "false"), resource.TestCheckResourceAttr(addr, "is_exclusive", "false"), resource.TestCheckResourceAttr(addr, "tags_ids.#", "0"), resource.TestCheckResourceAttrSet(addr, "namespace_id"), resource.TestCheckResourceAttrSet(addr, "created"),
		)},
		{Config: testAccIPAddressRangeConfig(seed, cidr, start, end, fmt.Sprintf("tfacc-range-updated-%d", seed), "updated by terraform", true), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(addr, "description", "updated by terraform"), resource.TestCheckResourceAttr(addr, "count_as_utilized", "true"))},
		{ResourceName: addr, ImportState: true, ImportStateVerify: true},
		{Config: testAccProviderConfig()},
	}})
}
