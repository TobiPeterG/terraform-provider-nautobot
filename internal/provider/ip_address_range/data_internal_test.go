package ip_address_range

import (
	"testing"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestIPAddressRangeModelFromAPIRejectsMalformedResponses(t *testing.T) {
	t.Parallel()
	if _, err := ipAddressRangeModelFromAPI(nil); err == nil {
		t.Fatal("expected nil response to fail")
	}
	if _, err := ipAddressRangeModelFromAPI(&nb.IPAddressRange{}); err == nil {
		t.Fatal("expected response without ID to fail")
	}
}
