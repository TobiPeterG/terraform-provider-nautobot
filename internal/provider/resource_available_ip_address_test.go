package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAvailableIPAddressConfigMinimal() string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_available_ip_address" "test" {
  prefix_id = "%s"
  status    = "%s"
}
`, testPrefixID, testStatus)
}

func testAccAvailableIPAddressConfigWithDNSName(dnsName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_available_ip_address" "test" {
  prefix_id = "%s"
  status    = "%s"
  dns_name  = "%s"
}
`, testPrefixID, testStatus, dnsName)
}

func testAccAvailableIPAddressConfigWithStatusAndDNSName(status, dnsName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_available_ip_address" "test" {
  prefix_id = "%s"
  status    = "%s"
  dns_name  = "%s"
}
`, testPrefixID, status, dnsName)
}

func testAccAvailableIPAddressConfigParallel() string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_available_ip_address" "ip1" {
  prefix_id = "%s"
  status    = "%s"
  dns_name  = "tfacc-parallel-1"
}

resource "nautobot_available_ip_address" "ip2" {
  prefix_id = "%s"
  status    = "%s"
  dns_name  = "tfacc-parallel-2"
}

resource "nautobot_available_ip_address" "ip3" {
  prefix_id = "%s"
  status    = "%s"
  dns_name  = "tfacc-parallel-3"
}
`, testPrefixID, testStatus,
		testPrefixID, testStatus,
		testPrefixID, testStatus)
}

func TestAccAvailableIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttrSet(resourceName, "address"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_version"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_update(t *testing.T) {
	t.Parallel()

	resourceName := "nautobot_available_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-ip-%d.example.com", time.Now().Unix())
	dnsName2 := fmt.Sprintf("tfacc-ip-upd-%d.example.com", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithDNSName(dnsName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName1),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName("Reserved", dnsName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName2),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName("Reserved", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_import(t *testing.T) {
	t.Parallel()

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_parallelAllocations(t *testing.T) {
	t.Parallel()

	resourceName1 := "nautobot_available_ip_address.ip1"
	resourceName2 := "nautobot_available_ip_address.ip2"
	resourceName3 := "nautobot_available_ip_address.ip3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigParallel(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "address"),
					resource.TestCheckResourceAttr(resourceName1, "status", testStatus),

					resource.TestCheckResourceAttr(resourceName2, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "address"),
					resource.TestCheckResourceAttr(resourceName2, "status", testStatus),

					resource.TestCheckResourceAttr(resourceName3, "prefix_id", testPrefixID),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "address"),
					resource.TestCheckResourceAttr(resourceName3, "status", testStatus),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
