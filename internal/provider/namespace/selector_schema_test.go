package namespace_test

import (
	"testing"

	namespacepkg "github.com/nautobot/terraform-provider-nautobot/internal/provider/namespace"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestNamespaceSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, namespacepkg.NewNamespaceDataSource(), "name")
}
