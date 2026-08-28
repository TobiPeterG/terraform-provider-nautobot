package prefix_test

import (
	"testing"

	prefixpkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/prefix"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestPrefixSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, prefixpkg.NewPrefixDataSource(), "prefix", "namespace_id")
}
