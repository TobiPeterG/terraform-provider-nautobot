package tenant_group_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/tenant_group"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestTenantGroupSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, tenant_group.NewTenantGroupDataSource(), "name")
}
