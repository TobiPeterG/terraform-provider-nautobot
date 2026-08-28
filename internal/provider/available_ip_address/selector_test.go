package available_ip_address_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider/testutil"
)

func TestAccAvailableIPAddressDataSource_missingSource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + `
data "nautobot_available_ip_address" "test" {}
`,
			ExpectError: regexp.MustCompile(`No attribute specified when one \(and only one\)`),
		}},
	})
}

func TestAccAvailableIPAddressDataSource_multipleSources(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.AccPreCheck(t) },
		ProtoV6ProviderFactories: testutil.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: testutil.AccProviderConfig() + `
data "nautobot_available_ip_address" "test" {
  prefix_id           = "ee995f43-0052-49f0-b0f4-e09fd47ca32f"
  ip_address_range_id = "bc18cc5c-d359-4942-ae36-39e2d525f908"
}
`,
			ExpectError: regexp.MustCompile(`2 attributes specified when one \(and only one\)`),
		}},
	})
}
