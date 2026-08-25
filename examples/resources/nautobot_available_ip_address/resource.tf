data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

// And finally get the first available IP address from that prefix
resource "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
  status    = "Active"
  dns_name  = "test-vm.test.com"
}

# An IP may alternatively be allocated from a non-exclusive IP address range.
resource "nautobot_available_ip_address" "from_range" {
  ip_address_range_id = "00000000-0000-0000-0000-000000000000"
  status              = "Active"
}

output "resource_available_ip_address" {
  value = nautobot_available_ip_address.example
}
