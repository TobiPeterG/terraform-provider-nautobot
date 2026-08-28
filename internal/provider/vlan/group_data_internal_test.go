package vlan

import (
	"testing"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestVLANGroupDataFromAPIRejectsMalformedResponses(t *testing.T) {
	t.Parallel()
	if _, err := vlanGroupDataFromAPI(nil); err == nil {
		t.Fatal("expected nil response to fail")
	}
	if _, err := vlanGroupDataFromAPI(&nb.VLANGroup{}); err == nil {
		t.Fatal("expected response without ID to fail")
	}
}
