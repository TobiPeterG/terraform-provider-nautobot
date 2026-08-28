package tenant_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/tenant"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestTenantSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, tenant.NewTenantDataSource(), "name")
}
