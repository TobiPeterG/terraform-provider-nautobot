package cluster_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/cluster"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestClusterSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, cluster.NewClusterDataSource(), "name")
}
