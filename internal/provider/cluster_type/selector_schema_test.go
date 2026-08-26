package cluster_type_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/cluster_type"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestClusterTypeSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, cluster_type.NewClusterTypeDataSource(), "name")
}
