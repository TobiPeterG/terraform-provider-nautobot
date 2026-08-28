package ip_address_range_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/ip_address_range"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestIPAddressRangeSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, ip_address_range.NewIPAddressRangeDataSource(), "start_address", "end_address", "namespace_id")
}
