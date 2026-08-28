package vlan_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/vlan"
)

func TestVLANSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, vlan.NewVLANDataSource(), "name", "vlan_group_id")
}

func TestVLANGroupSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, vlan.NewVLANGroupDataSource(), "name")
}
