package manufacturer_test

import (
	"testing"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/manufacturer"
	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestManufacturerSelectorSchema(t *testing.T) {
	testutil.AssertSingularSelectorSchema(t, manufacturer.NewManufacturerDataSource(), "name")
}
